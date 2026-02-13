package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

type Config struct {
	MinecraftVersion string `json:"minecraft_version"`
	MinRAM           string `json:"min_ram"`
	MaxRAM           string `json:"max_ram"`
	ServerDir        string `json:"server_dir"`
	JarName          string `json:"jar_name"`
	ServerURL        string `json:"server_url"`
	ServerIP         string `json:"server_ip"`
}

type fabricLoaderVersion struct {
	Version string `json:"version"`
	Stable  bool   `json:"stable"`
}

type fabricInstallerVersion struct {
	Version string `json:"version"`
	Stable  bool   `json:"stable"`
}

func main() {
	cfg, created, err := loadOrCreateConfig("launcher.json")
	if err != nil {
		fmt.Println("Не удалось прочитать конфиг:", err)
		os.Exit(1)
	}

	if created {
		fmt.Println("Создан launcher.json (Fabric 1.21.11, 5GB RAM). При необходимости измените параметры.")
	}

	if err := os.MkdirAll(cfg.ServerDir, 0o755); err != nil {
		fmt.Println("Не удалось создать папку сервера:", err)
		os.Exit(1)
	}

	if err := os.MkdirAll(filepath.Join(cfg.ServerDir, "mods"), 0o755); err != nil {
		fmt.Println("Не удалось создать папку mods:", err)
		os.Exit(1)
	}

	jarPath := filepath.Join(cfg.ServerDir, cfg.JarName)
	if _, err := os.Stat(jarPath); errors.Is(err, os.ErrNotExist) {
		fmt.Println("Fabric server launcher не найден. Скачиваю...")
		if err := downloadServerJar(cfg, jarPath); err != nil {
			fmt.Println("Не удалось скачать сервер:", err)
			os.Exit(1)
		}
	}

	if err := ensureEULA(cfg.ServerDir); err != nil {
		fmt.Println("Не удалось создать eula.txt:", err)
		os.Exit(1)
	}

	if err := writeServerIPFile("SERVER_IP.txt", cfg.ServerIP); err != nil {
		fmt.Println("Не удалось создать SERVER_IP.txt:", err)
		os.Exit(1)
	}

	if err := runServer(cfg); err != nil {
		fmt.Println("Ошибка запуска сервера:", err)
		os.Exit(1)
	}
}

func loadOrCreateConfig(path string) (Config, bool, error) {
	defaults := Config{
		MinecraftVersion: "1.21.11",
		MinRAM:           "5G",
		MaxRAM:           "5G",
		ServerDir:        "mc_server",
		JarName:          "fabric-server-launch.jar",
		ServerURL:        "",
		ServerIP:         "212.0.213.86",
	}

	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		data, err := json.MarshalIndent(defaults, "", "  ")
		if err != nil {
			return Config{}, false, err
		}
		if err := os.WriteFile(path, data, 0o644); err != nil {
			return Config{}, false, err
		}
		return defaults, true, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, false, err
	}

	cfg := defaults
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, false, err
	}

	if cfg.ServerDir == "" {
		cfg.ServerDir = defaults.ServerDir
	}
	if cfg.JarName == "" {
		cfg.JarName = defaults.JarName
	}
	if cfg.MinRAM == "" {
		cfg.MinRAM = defaults.MinRAM
	}
	if cfg.MaxRAM == "" {
		cfg.MaxRAM = defaults.MaxRAM
	}
	if cfg.MinecraftVersion == "" {
		cfg.MinecraftVersion = defaults.MinecraftVersion
	}
	if cfg.ServerIP == "" {
		cfg.ServerIP = defaults.ServerIP
	}

	return cfg, false, nil
}

func ensureEULA(serverDir string) error {
	eulaPath := filepath.Join(serverDir, "eula.txt")
	return os.WriteFile(eulaPath, []byte("eula=true\n"), 0o644)
}

func writeServerIPFile(path, ip string) error {
	content := fmt.Sprintf("IP сервера: %s\nПорт по умолчанию: 25565\n", ip)
	return os.WriteFile(path, []byte(content), 0o644)
}

func downloadServerJar(cfg Config, jarPath string) error {
	url := cfg.ServerURL
	if url == "" {
		buildURL, err := resolveFabricServerURL(cfg.MinecraftVersion)
		if err != nil {
			return err
		}
		url = buildURL
	}

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("неожиданный HTTP статус: %s", resp.Status)
	}

	out, err := os.Create(jarPath)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return err
	}

	return nil
}

func resolveFabricServerURL(mcVersion string) (string, error) {
	loaderVersion, err := resolveLatestStableLoaderVersion()
	if err != nil {
		return "", err
	}

	installerVersion, err := resolveLatestStableInstallerVersion()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(
		"https://meta.fabricmc.net/v2/versions/loader/%s/%s/%s/server/jar",
		mcVersion,
		loaderVersion,
		installerVersion,
	), nil
}

func resolveLatestStableLoaderVersion() (string, error) {
	resp, err := http.Get("https://meta.fabricmc.net/v2/versions/loader")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Fabric loader API вернуло %s", resp.Status)
	}

	var versions []fabricLoaderVersion
	if err := json.NewDecoder(resp.Body).Decode(&versions); err != nil {
		return "", err
	}

	for _, v := range versions {
		if v.Stable {
			return v.Version, nil
		}
	}

	if len(versions) == 0 {
		return "", errors.New("Fabric loader версии не найдены")
	}

	return versions[0].Version, nil
}

func resolveLatestStableInstallerVersion() (string, error) {
	resp, err := http.Get("https://meta.fabricmc.net/v2/versions/installer")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Fabric installer API вернуло %s", resp.Status)
	}

	var versions []fabricInstallerVersion
	if err := json.NewDecoder(resp.Body).Decode(&versions); err != nil {
		return "", err
	}

	for _, v := range versions {
		if v.Stable {
			return v.Version, nil
		}
	}

	if len(versions) == 0 {
		return "", errors.New("Fabric installer версии не найдены")
	}

	return versions[0].Version, nil
}

func runServer(cfg Config) error {
	javaPath, err := exec.LookPath("java")
	if err != nil {
		return fmt.Errorf("java не найдена в PATH. Установите Java 21+ или добавьте путь к java.exe в PATH")
	}

	cmd := exec.Command(javaPath,
		"-Xms"+cfg.MinRAM,
		"-Xmx"+cfg.MaxRAM,
		"-jar",
		cfg.JarName,
		"nogui",
	)
	cmd.Dir = cfg.ServerDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	fmt.Println("Запускаю Minecraft Fabric сервер...")
	fmt.Printf("Папка: %s\n", cfg.ServerDir)
	fmt.Printf("Память: -Xms%s -Xmx%s\n", cfg.MinRAM, cfg.MaxRAM)
	fmt.Printf("IP для друзей: %s\n", cfg.ServerIP)
	fmt.Println("Моды можно добавлять в папку mc_server/mods (jar-файлы).")

	return cmd.Run()
}
