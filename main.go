package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

/*
Перевірка чи існує файл
*/
func ifExistFile(filename string) bool {
	if _, err := os.Stat(filename); err != nil {
		file, err2 := os.Create(filename)
		if err2 != nil {
			log.Fatalf("Виникла помилка під час створення файлу %s. \n", filename)
			return false
		}
		addHeader(file) //додати хедер до файлу
		fmt.Printf("Файл %s не існує, тому його було створено\n", filename)
		closeFile(file)
		return true
	}
	return true
}

/*
Закриття файлу
*/
func closeFile(file *os.File) {
	if err := file.Close(); err != nil {
		log.Fatal(err)
	}
}

/*
Додання головної лінії (хедеру) в файл
*/
func addHeader(file *os.File) {
	write, err := file.WriteString("Surname, Group, Type of study (contract or budget)\n")
	if err != nil {
		closeFile(file)
		log.Fatalf("Не вдалося додати строку до файлу: %s \n %d", err, write)
	}
}

/*
Додання строки в файл
*/
func appendToFile(filename string, yourString string) bool {

	// Відкриття файлу для додавання строк в кінець файлу
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		closeFile(file)
		log.Fatalf("Не вдалося відкрити файл %s: %s", filename, err)
		return false
	}

	// Запис строки в файл
	write, err := file.WriteString(yourString + "\n")
	if err != nil {
		closeFile(file)
		log.Fatalf("Не вдалося додати строку до файлу %s: %s \n %d", filename, err, write)
		return false
	}
	closeFile(file)

	return true
}

/*
Сортування строк
*/
func sortLines(inputFilename string, outputFilename string, revers bool, sortByColumnNumber int, ignoreHeader bool) bool {
	var sorted sort.StringSlice
	var tempString string
	//var i int

	if ifExistFile(outputFilename) {
		// Відкриття файлу
		file, err := os.OpenFile(outputFilename, os.O_RDONLY, 0644)
		if err != nil {
			log.Fatalf("Не вдалося відкрити файл %s: %s", inputFilename, err)
			return false
		}
		closeFile(file)

		tempString = oF(inputFilename, ignoreHeader)

		sorted = strings.Split(tempString, "\n") // конвертуємо строку в Slice з роздільником \n

		if revers {
			//Якщо необхідно сортувати в зворотньому порядку, сортуємо в зворотньому порядку
			sorted.Sort()
			sort.Sort(sort.Reverse(sorted))
		} else {
			sorted.Sort()
		}
		pushStringsToFile(sorted, outputFilename, ignoreHeader)
		return true
	} else {
		return false
	}

}

/*
Ітерація файлів у папці
*/
func iterate(path string) []string {
	var files []string

	//Проходимося по кожній папці та зчитуємо файли .csv
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatalf(err.Error())
		}
		if !info.IsDir() && filepath.Ext(path) == ".csv" {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil
	}
	return files
}

/*
Занести строки в файл
*/
func pushStringsToFile(sorted sort.StringSlice, outputFilename string, ignoreHeader bool) bool {

	//Відкриваємо файл, в який необхідно додати строки
	file, err := os.OpenFile(outputFilename, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("Виникла помилка під час створення файлу: %s", err)
	}
	//Очищуємо файл від попереднього змісту
	if err := os.Truncate(outputFilename, 0); err != nil {
		log.Printf("Failed to truncate: %v", err)
	}
	//Додаємо Хедер, якщо задано флаг -h
	if ignoreHeader {
		addHeader(file)
	}
	//Записуємо кожну строку в файл
	datawriter := bufio.NewWriter(file)
	for _, data := range sorted {
		if len(data) != 0 {
			_, _ = datawriter.WriteString(data + "\n")
		}
	}
	//Посилаємо текст з буферу в файл
	err = datawriter.Flush()
	if err != nil {
		log.Fatalf("Виникла помилка під час запису відсортованих даних в файл. %s", err)
		return false
	}

	closeFile(file)
	return true
}

/*
OpenFile функція. Читання строк
*/
func oF(filename string, ignoreHeader bool) string {

	var tempString string
	fileIn, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Не вдалося відкрити файл %s: %s", filename, err)
		return ""
	}
	//Читання кожної строки файлу та формуємо їх в одну строку
	reader := bufio.NewReader(fileIn)
	i := 0

	for {
		line, _, erro := reader.ReadLine()

		// Якщо строки завершилися, перериваємо цикл читання
		if erro == io.EOF {
			break
		}

		//Якщо потрібно ігнорувати першу строку, ігноруємо її
		if ignoreHeader {
			if i != 0 {
				tempString += string(line) + "\n"
			}
		} else {
			tempString += string(line) + "\n"
		}
		i++
	}
	//Закриття файлу
	closeFile(fileIn)
	return tempString
}

func sortLinesFromManyFiles(files []string, outputFileName string, revers bool, sortByColumnNumber int, ignoreHeader bool) bool {

	var sorted sort.StringSlice
	var tempString string

	readString := make(chan string) //відкриття каналу для занесення строк

	for j := 0; j < len(files); j++ {
		// Створення goroutineS
		go func(pathOfFile string) {
			readString <- oF(pathOfFile, ignoreHeader) //відправка даних в канал
		}(files[j])

		tempString += <-readString //читання даних з каналу
	}

	sorted = strings.Split(tempString, "\n") // конвертуємо строку в Slice з роздільником \n

	if revers {
		//Якщо необхідно сортувати в зворотньому порядку, сортуємо в зворотньому порядку
		sorted.Sort()
		sort.Sort(sort.Reverse(sorted))
	} else {
		sorted.Sort()
	}
	pushStringsToFile(sorted, outputFileName, ignoreHeader)
	return true

}

/*
Розпочати програму
*/
func startProgram(inputFile string, outputFile string, f int, h bool, r bool, dirName string) bool {
	for { //нескінченний цикл
		var nameSt, groupSt, typeSt string //змінні з ім'ям, групою, видом навчання (контракт/бюджет)
		var value int                      //для вибору щодо додавання нової строки
		var fileExist, appendedString bool //існує файл filename (true false)
		var listOfFiles []string

		if dirName != "/" {
			//Якщо задано ім'я папки, зчитуємо шлях поточної директорії
			currentDirectory, err := os.Getwd()
			if err != nil {
				log.Fatal(err)
			}

			// Зчитуємо список файлів .csv
			listOfFiles = iterate(currentDirectory + "/" + dirName)
			// Сортуємо строки
			sortLinesFromManyFiles(listOfFiles, outputFile, r, 0, h)
			return true
		} else {
			// Якщо не задано ключ -d
			fmt.Println("Введіть строку в форматі \"Прізвище Група Вид_навчання(К чи Б)\":")
			_, err := fmt.Scanln(&nameSt, &groupSt, &typeSt) //Введення строки
			if err != nil {
				log.Fatalln("Виникла помилка - ", err)
			}

			fileExist = ifExistFile(inputFile) //перевірка на існування файлу
			if fileExist {
				appendedString = appendToFile(inputFile, nameSt+","+groupSt+","+typeSt) //додання строки в файл
				if appendedString {
					fmt.Println("Успішно додано строку в файл!")
					fmt.Printf("Бажаєте додати ще строку? 1 - так, 0 - ні: ")
					_, err = fmt.Scanln(&value)
					if err != nil {
						sortLines(inputFile, outputFile, r, 0, h)
						log.Fatalf("Виникла помилка: %s", err)
						return false
					}
					if value != 1 {
						sortLines(inputFile, outputFile, r, 0, h)
						fmt.Printf("Вихід з програми.")
						return true
					}
				} else {
					fmt.Printf("Не вдалося додати строку в файл %s.", inputFile)
					return false
				}

			} else {
				fmt.Printf("Не вдалося створити файл %s.", inputFile)
				return false
			}
		}

	}

}

func main() {
	var inputFile, outputFile, dirName string
	var number int
	var h, r bool

	flag.Bool("help", false, "Отримати довідку")
	flag.BoolVar(&h, "h", false, "Ігнорувати першу строку під час сортування даних")
	flag.StringVar(&inputFile, "i", "test.csv", "Використовувати файл file-name як вхідний файл")
	flag.StringVar(&outputFile, "o", "test.csv", "Використовувати файл file-name як результуючий файл")
	flag.IntVar(&number, "f", 0, "Сортувати вхідні строки за стовпцем N")
	flag.StringVar(&dirName, "d", "/", "dir-name - Задати вхідну папку")
	flag.BoolVar(&r, "r", false, "Реверсне сортування даних")
	flag.Parse()

	if outputFile != "test.csv" {
		fmt.Printf("Обрано результуючий файл %s\n", outputFile)

		if _, err := os.Stat(outputFile); os.IsNotExist(err) {
			if err := os.MkdirAll(filepath.Dir(outputFile), 0770); err != nil {
				log.Fatalln("Помилка")
			}
			_, err := os.Create(outputFile)
			if err != nil {
				return
			}

		}
	}
	if inputFile != "test.csv" {
		fmt.Printf("Обрано вхідний файл %s. Подальші операції виконуватимуться з ним.\n", outputFile)
		if _, err := os.Stat(outputFile); os.IsNotExist(err) {
			if err := os.MkdirAll(filepath.Dir(outputFile), 0770); err != nil {
				log.Fatalln("Помилка")
			}
			_, err := os.Create(outputFile)
			if err != nil {
				return
			}

		}
	}
	//if number != 0 {
	//	fmt.Printf("Виконувати сортування строк за стовпцем №%d\n", number)
	//}
	if r != false {
		fmt.Printf("Виконувати реверсне сортування даних\n")
	}
	if h != false {
		fmt.Printf("Ігнорувати хедер під час сортування даних\n")
	}
	if dirName != "/" {
		fmt.Printf("Задано вхідну папку %s\n", dirName)
	}
	if dirName != "/" && inputFile != "test.csv" {
		log.Fatalln("Помилка! Необхідно задати лише один з параметрів -d або -i.")
	}
	startProgram(inputFile, outputFile, number, h, r, dirName)

}
