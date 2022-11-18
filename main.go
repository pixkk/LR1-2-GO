package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
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
	var i int

	if ifExistFile(outputFilename) {
		// Відкриття файлу
		file, err := os.OpenFile(outputFilename, os.O_RDONLY, 0644)
		if err != nil {
			log.Fatalf("Не вдалося відкрити файл %s: %s", inputFilename, err)
			return false
		}
		closeFile(file)

		fileIn, err := os.Open(inputFilename)
		if err != nil {
			log.Fatalf("Не вдалося відкрити файл %s: %s", outputFilename, err)
			return false
		}
		//Читання кожної строки файлу та формуємо їх в одну строку
		reader := bufio.NewReader(fileIn)
		i = 0
		for {
			line, _, err := reader.ReadLine()

			// Якщо строки завершилися, перериваємо цикл читання
			if err == io.EOF {
				break
			}

			//Якщо потрібно ігнорувати першу строку, ігноруємо її
			if ignoreHeader {
				if i != 0 {
					tempString = tempString + string(line) + "\n"
				}
			} else {
				tempString = tempString + string(line) + "\n"
			}
			i++
		}
		//Закриття файлу
		closeFile(fileIn)

		sorted = strings.Split(tempString, "\n") // конвертуємо строку в Slice з роздільником \n

		if revers {
			//Якщо необхідно сортувати в зворотньому порядку, сортуємо в зворотньому порядку
			sorted.Sort()
			sort.Sort(sort.Reverse(sorted))
		} else {
			sorted.Sort()
		}

		//Відкриваємо файл, в який необхідно додати строки
		file, err = os.OpenFile(outputFilename, os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			log.Fatalf("Виникла помилка під час створення файлу: %s", err)
		}
		//Очищуємо файл від попереднього змісту
		if err := os.Truncate(outputFilename, 0); err != nil {
			log.Printf("Failed to truncate: %v", err)
		}
		//Додаємо Хедер, якщо задано флаг -r
		if ignoreHeader {
			addHeader(file)
		}

		//Записуємо кожну строку в файл
		datawriter := bufio.NewWriter(file)
		for _, data := range sorted {
			if len(data) != 0 {
				fmt.Println("Resulte:", data)
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
	} else {
		return false
	}

}

/*
Розпочати програму
*/
func startProgram(inputFile string, outputFile string, f int, h bool, r bool) bool {
	for { //нескінченний цикл

		var nameSt, groupSt, typeSt string //змінні з ім'ям, групою, видом навчання (контракт/бюджет)
		var value int                      //для вибору щодо додавання нової строки
		var fileExist, appendedString bool //існує файл filename (true false)

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

func main() {
	var inputFile, outputFile string
	var number int
	var h, r bool
	flag.Bool("help", false, "Отримати довідку")
	flag.BoolVar(&h, "h", false, "Ігнорувати першу строку під час сортування даних")
	flag.StringVar(&inputFile, "i", "test.csv", "Використовувати файл file-name як вхідний файл")
	flag.StringVar(&outputFile, "o", "test.csv", "Використовувати файл file-name як результуючий файл")
	flag.IntVar(&number, "f", 0, "Сортувати вхідні строки за стовпцем N")
	flag.BoolVar(&r, "r", false, "Реверсне сортування даних")
	flag.Parse()

	if outputFile != "test.csv" {
		fmt.Printf("Обрано результуючий файл %s\n", outputFile)
	}
	if inputFile != "test.csv" {
		fmt.Printf("Обрано вхідний файл %s. Подальші операції виконуватимуться з ним.\n", outputFile)
	}
	if number != 0 {
		fmt.Printf("Виконувати сортування строк за стовпцем №%d\n", number)
	}
	if r != false {
		fmt.Printf("Виконувати реверсне сортування даних\n")
	}
	if h != false {
		fmt.Printf("Ігнорувати хедер під час сортування даних\n")
	}
	startProgram(inputFile, outputFile, number, h, r)
}
