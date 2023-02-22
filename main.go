package main

import (
	"encoding/json"
	"fmt"
	"golang.org/x/exp/slices"
	"log"
	"os"
	"os/exec"
	"path"
	"regexp"
	"sort"
	"strings"
)

// composerStruct 是php项目composer.json的解析映射
// 主要的是php部分确定php版本
type composerStruct struct {
	Require struct {
		Php string `json:"php,omitempty,-"`
	} `json:"require,omitempty,-"`
}

func main() {

	filepath, content := loadConfigurations()

	if content == "" {
		log.Fatalf("!!! EMPTY CONFIG [%s] !!!", filepath)
	}

	configs := parseIni(content)

	if _, ok := configs["default"]; !ok {
		log.Fatalln("!!! NOT FOUND [default] CONFIG ITEMS !!!")
	}

	// 默认要使用的php版本
	def := configs["default"]

	if _, ok := configs[def]; !ok {
		log.Fatalln("!!! [default] PHP version not found in config items. !!!")
	}

	// 默认要使用的php版本
	bin := configs[configs["default"]]
	ver := ""

	// 获取目标php版本
	targetVersion, err := getTargetVersion()
	if err != nil {
		targetVersion = configs["default"]
	}
	// 删除default的配置
	delete(configs, "default")

	// 统计版本按顺序排序
	versions := make([]string, 0)
	for v, _ := range configs {
		if v == "default" {
			continue
		}
		versions = append(versions, v)
	}
	slices.Sort(versions)

	opt, target, err := getOperator(targetVersion)
	if err != nil {
		log.Fatalln(err)
	}

	for _, base := range versions {
		if opt.Compare(base, target) {
			ver = base
			break
		}
	}

	if ver == "" {
		log.Fatalln("detect fail")
	}

	bin = configs[ver]

	if bin == "" {
		log.Fatalln("!!! invalid version detect !!!")
	}

	cmd := exec.Command(bin, os.Args[1:]...)
	cmd.Env = os.Environ()
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		log.Fatalln(err)
	}

}

// getTargetVersion 获取当前需要执行的PHP版本
// 优先级 ENV > phpver > composer.json
func getTargetVersion() (string, error) {

	// 从环境变量读取php版本
	// 优先级最高
	// PV=7.2 php xxx
	PV := os.Getenv("PV")

	if PV != "" {
		return PV, nil
	}

	if cwd, err := os.Getwd(); err == nil {
		// 检查当前路径有无phpver文件
		// 如果有则从当前路径读取phpver文件并获取版本
		PHPVersionFile := path.Join(cwd, "phpver")
		if fileExists(PHPVersionFile) {
			PHPVersionFileContent := readFile(PHPVersionFile)
			if PHPVersionFileContent == "default" {
				return "", fmt.Errorf("!!! invalid [php version file] sets ")
			}
			return PHPVersionFileContent, nil
		}

		// 检查当前路径有无composer.json文件
		// 如果有则从当前路径读取composer.json文件并获取版本
		ComposerJSONFILE := "composer.json"
		if fileExists(ComposerJSONFILE) {
			ComposerJSONFileContent := readFile(ComposerJSONFILE)
			var composerJsonStruct composerStruct
			err := json.Unmarshal([]byte(ComposerJSONFileContent), &composerJsonStruct)
			if err != nil {
				return "", err
			}
			if composerJsonStruct.Require.Php != "" {
				// 获取composer内执行的php依赖版本
				var composerVersions Vers
				composerVersions = strings.Split(composerJsonStruct.Require.Php, "|")
				if len(composerVersions) == 1 {
					return composerVersions[0], nil
				}
				sort.Sort(composerVersions)
				return composerVersions[0], nil
			}
		}
	}

	return "", fmt.Errorf("parse target version fail")
}

func loadConfigurations() (string, string) {
	cfgDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalln(err)
		return "", ""
	}

	cfgDir = cfgDir + "/.config"
	if !directoryExists(cfgDir) {
		log.Printf("%s not found\n", cfgDir+"/.config")
		return "", ""
	}

	cfgFile := cfgDir + "/php-detect"
	if !fileExists(cfgFile) {
		if err = writeFile(cfgFile, ""); err != nil {
			log.Fatalln(err)
			return cfgFile, ""
		}
	}

	return cfgFile, readFile(cfgFile)
}

func readFile(filepath string) string {

	file, err := os.ReadFile(filepath)
	if err != nil {
		log.Fatalf("readfile err: %s\n", err)
		return ""
	}

	return string(file)

}

func writeFile(filepath string, content string) error {
	return os.WriteFile(filepath, []byte(content), 0644)
}

func directoryExists(filepath string) bool {
	stat, err := os.Stat(filepath)

	if os.IsNotExist(err) {
		return false
	}

	if !stat.IsDir() {
		return false
	}

	return true
}

func fileExists(filepath string) bool {
	_, err := os.Stat(filepath)

	if os.IsNotExist(err) {
		return false
	}

	return true
}

func parseIni(content string) map[string]string {
	ini := make(map[string]string)

	keyValue := regexp.MustCompile(`^(\w.*)\s*=\s*(.*?)\s*$`)

	contents := strings.Split(strings.ReplaceAll(content, "\r", ""), "\n")

	for _, line := range contents {

		if line == "" {
			continue
		}

		result := keyValue.FindStringSubmatch(line)
		if len(result) > 0 {
			key, value := result[1], result[2]
			// fmt.Printf("kv: '%q'\n", result)
			ini[key] = value
			continue
		}
	}

	return ini
}

func getOperator(targetVersion string) (Operator, string, error) {
	for _, opt := range Operators {
		if strings.HasPrefix(targetVersion, opt.Symbol()) {
			return opt, strings.TrimPrefix(targetVersion, opt.Symbol()), nil
		}
	}

	return &Equal{&operator{"="}}, targetVersion, nil
}
