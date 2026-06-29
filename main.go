package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/EvilMilko/Obfuscator-1C/obfuscator"
)

// AppConfig зеркалирует параметры оригинальной библиотеки obfuscator.Config
type AppConfig struct {
	RepExpByTernary  bool     `json:"RepExpByTernary"`
	RepLoopByGoto    bool     `json:"RepLoopByGoto"`
	RepExpByEval     bool     `json:"RepExpByEval"`
	HideString       bool     `json:"HideString"`
	ChangeConditions bool     `json:"ChangeConditions"`
	AppendGarbage    bool     `json:"AppendGarbage"`
	CallStackHell    bool     `json:"CallStackHell"`
	LineBreaks       bool     `json:"LineBreaks"`
	NoEvalFuncs      []string `json:"NoEvalFuncs"`
}

// ToObfuscatorConfig преобразует AppConfig в конфигурационную структуру библиотеки
func (c AppConfig) ToObfuscatorConfig() obfuscator.Config {
	return obfuscator.Config{
		RepExpByTernary:  c.RepExpByTernary,
		RepLoopByGoto:    c.RepLoopByGoto,
		RepExpByEval:     c.RepExpByEval,
		HideString:       c.HideString,
		ChangeConditions: c.ChangeConditions,
		AppendGarbage:    c.AppendGarbage,
		CallStackHell:    c.CallStackHell,
		LineBreaks:       c.LineBreaks,
		NoEvalFuncs:      c.NoEvalFuncs,
	}
}

// loadConfig выполняет загрузку настроек из config.json или возвращает безопасные значения по умолчанию
func loadConfig(path string) AppConfig {
	// Безопасные дефолтные значения
	config := AppConfig{
		HideString: true,
		LineBreaks: true,
	}

	file, err := os.Open(path)
	if err != nil {
		log.Printf("Предупреждение: файл настроек %s не найден (%v). Использование безопасных параметров по умолчанию (HideString=true, LineBreaks=true).", path, err)
		return config
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		log.Printf("Предупреждение: ошибка парсинга JSON в файле %s (%v). Использование параметров по умолчанию.", path, err)
	}

	return config
}

// copyFile копирует файл байт-в-байт
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// processBslFile выполняет обфускацию bsl-файла с обработкой BOM, пустых файлов и защитой от паники
func processBslFile(src, dst string, config obfuscator.Config, successObfuscated, copiedAsIs, skippedBsl, errorCount *int) error {
	data, err := os.ReadFile(src)
	if err != nil {
		*errorCount++
		*skippedBsl++
		return fmt.Errorf("ошибка чтения файла: %w", err)
	}

	// Проверка на UTF-8 BOM (\xef\xbb\xbf)
	hasBOM := false
	content := data
	if len(data) >= 3 && data[0] == 0xef && data[1] == 0xbb && data[2] == 0xbf {
		hasBOM = true
		content = data[3:]
	}

	// Проверка на пустой файл
	if len(content) == 0 {
		if err := os.WriteFile(dst, data, 0644); err != nil {
			*errorCount++
			*skippedBsl++
			return fmt.Errorf("ошибка записи пустого файла: %w", err)
		}
		*copiedAsIs++
		return nil
	}

	var obfuscatedCode string
	var obfErr error
	var panicked bool

	// Защита от сбоев и паники при вызове библиотеки обфускатора
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicked = true
				obfErr = fmt.Errorf("паника библиотеки обфускатора: %v", r)
			}
		}()

		obf := obfuscator.NewObfuscatory(context.Background(), config)
		obfuscatedCode, obfErr = obf.Obfuscate(string(content))
	}()

	if panicked || obfErr != nil {
		log.Printf("Предупреждение: сбой обфускации файла %s: %v. Копирование оригинала.", src, obfErr)
		if err := os.WriteFile(dst, data, 0644); err != nil {
			*errorCount++
			*skippedBsl++
			return fmt.Errorf("ошибка сохранения исходного файла: %w", err)
		}
		*errorCount++
		*skippedBsl++
		return nil
	}

	// Сборка итогового содержимого
	var finalData []byte
	if hasBOM {
		finalData = append([]byte{0xef, 0xbb, 0xbf}, []byte(obfuscatedCode)...)
	} else {
		finalData = []byte(obfuscatedCode)
	}

	if err := os.WriteFile(dst, finalData, 0644); err != nil {
		*errorCount++
		*skippedBsl++
		return fmt.Errorf("ошибка записи обфусцированного файла: %w", err)
	}

	*successObfuscated++
	return nil
}

func main() {
	dirFlag := flag.String("dir", "", "Путь к каталогу выгрузки 1С (обязательный аргумент)")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Использование: 1c-obfuscator.exe -dir=<путь_к_каталогу>\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Утилита для рекурсивной обфускации исходных кодов 1С (файлов модулей .bsl).\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Создает копию каталога с суффиксом _obfuscated на том же уровне.\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Параметры командной строки:\n")
		flag.PrintDefaults()
		fmt.Fprintf(flag.CommandLine.Output(), "\nПараметры конфигурации в файле config.json:\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  RepExpByTernary (bool):\n\tЗаменять простые выражения тернарными операторами.\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  RepLoopByGoto (bool):\n\tЗаменять стандартные циклы на безусловные переходы с метками 'Перейти'.\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  RepExpByEval (bool):\n\tПрятать вызовы методов и выражения в платформенные функции 'Выполнить()' и 'Вычислить()'.\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  HideString (bool):\n\tШифровать и скрывать строковые константы (base64 + XOR с динамическим ключом).\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  ChangeConditions (bool):\n\tЗапутывать логические условия в операторах 'Если' добавлением фиктивных истинных условий по И.\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  AppendGarbage (bool):\n\tДобавлять неиспользуемый мусорный код (лишние переменные, пустые условные переходы).\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  CallStackHell (bool):\n\tПрятать выражения за цепочками вызовов случайно сгенерированных функций.\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  LineBreaks (bool):\n\tСохранять переносы строк (true) или сворачивать результирующий BSL-код в одну строку (false).\n")
	}
	flag.Parse()

	if *dirFlag == "" {
		log.Fatal("Ошибка: аргумент -dir является обязательным")
	}

	inputDir := filepath.Clean(*dirFlag)

	// Валидация входного пути
	info, err := os.Stat(inputDir)
	if err != nil {
		log.Fatalf("Ошибка: указанный путь не существует: %s", inputDir)
	}
	if !info.IsDir() {
		log.Fatalf("Ошибка: указанный путь не является директорией: %s", inputDir)
	}

	// Расчет имени выходной папки
	outputDir := inputDir + "_obfuscated"

	// Создание и очистка выходной папки
	log.Printf("Очистка и создание выходного каталога: %s", outputDir)
	if err := os.RemoveAll(outputDir); err != nil {
		log.Fatalf("Ошибка при очистке выходной папки: %v", err)
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Ошибка при создании выходной папки: %v", err)
	}

	// Чтение конфигурации
	config := loadConfig("config.json")
	obfConfig := config.ToObfuscatorConfig()

	// Счётчики статистики
	var totalFiles int
	var successObfuscated int
	var copiedAsIs int
	var skippedBsl int
	var errorCount int

	// Рекурсивный обход каталогов (Walker)
	err = filepath.WalkDir(inputDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			log.Printf("Ошибка при обработке пути %s: %v", path, err)
			errorCount++
			return nil
		}

		relPath, err := filepath.Rel(inputDir, path)
		if err != nil {
			log.Printf("Ошибка при получении относительного пути для %s: %v", path, err)
			errorCount++
			return nil
		}

		targetPath := filepath.Join(outputDir, relPath)

		if d.IsDir() {
			if path == inputDir {
				return nil
			}
			if err := os.MkdirAll(targetPath, 0755); err != nil {
				log.Printf("Ошибка при создании каталога %s: %v", targetPath, err)
				errorCount++
			}
			return nil
		}

		totalFiles++

		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".bsl" {
			// Копируем не-bsl файл байт-в-байт
			if err := copyFile(path, targetPath); err != nil {
				log.Printf("Ошибка при копировании файла %s: %v", path, err)
				errorCount++
			} else {
				copiedAsIs++
			}
			return nil
		}

		// Обработка bsl файла
		if err := processBslFile(path, targetPath, obfConfig, &successObfuscated, &copiedAsIs, &skippedBsl, &errorCount); err != nil {
			log.Printf("Ошибка обработки bsl-файла %s: %v", path, err)
		}

		return nil
	})

	if err != nil {
		log.Fatalf("Критическая ошибка при рекурсивном обходе: %v", err)
	}

	// Итоговое логирование
	fmt.Println()
	log.Println("=== Статистика работы ===")
	log.Printf("Всего файлов обработано: %d", totalFiles)
	log.Printf("Успешно обфусцировано:   %d", successObfuscated)
	log.Printf("Скопировано без изменений (не-bsl): %d", copiedAsIs)
	log.Printf("Пропущено bsl (пустые/ошибки):      %d", skippedBsl)
	log.Printf("Всего ошибок возникло:   %d", errorCount)
}
