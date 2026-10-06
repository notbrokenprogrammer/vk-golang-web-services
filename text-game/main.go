package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
)

var kitchen *Rooms
var room *Rooms
var corridor *Rooms
var street *Rooms
var student *Student

/*
код писать в этом файле
наверняка у вас будут какие-то структуры с методами, глобальные переменные ( тут можно ), функции
*/

type Student struct {
	inventory []string
	backpack  bool
	position  *Rooms
}

func (s *Student) put(item string) string {
	if s.backpack {
		s.inventory = append(s.inventory, item)
		result := fmt.Sprintf("предмет добавлен в инвентарь: %s", item)
		return result
	} else {
		return "некуда класть"
	}
}

type Kitchen struct {
	name  string
	items map[string]bool
}
type Rooms struct {
	name   string
	items  map[string]bool
	toPass []string
}

/* осмотреться, идти ..., надеть ..., взять ..., применить ..., завтракать ...,  */
func main() {
	/*
		в этой функции можно ничего не писать,
		но тогда у вас не будет работать через go run main.go
		очень круто будет сделать построчный ввод команд тут, хотя это и не требуется по заданию
	*/

	initGame()
	scaner := bufio.NewScanner(os.Stdin)

	for scaner.Scan() {
		text := scaner.Text()
		fmt.Println(handleCommand(text))

	}
	if err := scaner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "Ошибка чтения:", err)
	}
}

func initGame() {
	/*
		эта функция инициализирует игровой мир - все комнаты
		если что-то было - оно корректно перезатирается
	*/
	kitchen = &Rooms{
		name: "кухня",
		items: map[string]bool{
			"чай": true,
		},
		toPass: []string{"коридор"},
	}
	room = &Rooms{
		name: "комната",
		items: map[string]bool{
			"ключи":     true,
			"конспекты": true,
			"рюкзак":    true,
		},
		toPass: []string{"коридор"},
	}
	corridor = &Rooms{
		name: "коридор",
		items: map[string]bool{
			"дверь": false,
		},
		toPass: []string{"кухня", "комната", "улица"},
	}
	street = &Rooms{
		name:   "улица",
		items:  map[string]bool{},
		toPass: []string{"домой"},
	}
	student = &Student{
		inventory: []string{},
		backpack:  false,   // Надет рюкзак или нет
		position:  kitchen, // ожидает адрес объекта
	}

}

func handleCommand(command string) string {
	/*
		данная функция принимает команду от "пользователя"
		и наверняка вызывает какой-то другой метод или функцию у "мира" - списка комнат
	*/
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return "неизвестная команда"
	}
	switch parts[0] {
	case "осмотреться":
		if len(parts) != 1 {
			return "некорректная длина команды"
		}
		switch student.position.name {
		case "коридор":
			return "ничего интересного. можно пройти - " + strings.Join(student.position.toPass, ", ")
		case "улица":
			return "на улице весна. можно пройти - " + strings.Join(student.position.toPass, ", ")
		}

		tableItems := []string{}
		backpackOnChair := false

		for name, present := range student.position.items {
			if !present {
				continue
			}

			if name == "рюкзак" {
				backpackOnChair = !student.backpack
			} else {
				tableItems = append(tableItems, name)
			}
		}

		sort.Strings(tableItems)

		descriptionParts := []string{}
		isKitchen := student.position.name == "кухня"

		if isKitchen {
			descriptionParts = append(descriptionParts, "ты находишься на кухне")
		}

		if len(tableItems) > 0 {
			descriptionParts = append(descriptionParts, "на столе: "+strings.Join(tableItems, ", "))
		}

		if backpackOnChair {
			descriptionParts = append(descriptionParts, "на стуле: рюкзак")
		}

		if isKitchen {
			goal := "надо идти в универ"
			if !student.backpack {
				goal = "надо собрать рюкзак и идти в универ"
			}
			descriptionParts = append(descriptionParts, goal)
		}

		if len(descriptionParts) == 0 {
			descriptionParts = append(descriptionParts, "пустая комната")
		}

		roomDescription := strings.Join(descriptionParts, ", ")
		passagesDescription := strings.Join(student.position.toPass, ", ")

		return fmt.Sprintf("%s. можно пройти - %s", roomDescription, passagesDescription)
	case "идти": // Добавить смену адреса объекта
		if len(parts) != 2 {
			return "некорректная длина команды"
		}
		for _, v := range student.position.toPass {
			if parts[1] == v {
				switch v {
				case "коридор", "домой":
					student.position = corridor
					return "ничего интересного. можно пройти - кухня, комната, улица"
				case "комната":
					student.position = room
					return "ты в своей комнате. можно пройти - коридор"
				case "кухня":
					student.position = kitchen
					return "кухня, ничего интересного. можно пройти - коридор"
				case "улица":

					if student.position.items["дверь"] {
						student.position = street

						return "на улице весна. можно пройти - домой"
					} else {
						return "дверь закрыта"
					}
				default:
					result := fmt.Sprintf("нет пути в %s", v)
					return result
				}

			}
		}
		res := fmt.Sprintf("нет пути в %s", parts[1])
		return res

	case "надеть":
		if len(parts) != 2 {
			return "некорректная длина команды"
		}
		if parts[1] == "рюкзак" && student.position.name == "комната" && student.position.items["рюкзак"] {
			student.backpack = true
			student.position.items["рюкзак"] = false
			return "вы надели: рюкзак"
		} else {
			result := fmt.Sprintf("невозможно надеть: %s", parts[1])
			return result
		}
	case "применить":
		if len(parts) != 3 {
			return "некорректная длина команды"
		}
		flag := false
		for _, k := range student.inventory {
			if k == parts[1] {
				flag = true
				break
			}
		}
		if !flag {
			return fmt.Sprintf("нет предмета в инвентаре - %s", parts[1])
		} else if parts[1] == "ключи" && parts[2] == "дверь" {
			if _, exists := student.position.items["дверь"]; !exists {
				return "не к чему применить"
			}
			student.position.items["дверь"] = true // Открываем дверь
			return "дверь открыта"
		} else {
			return "не к чему применить"
		}

	case "взять":
		if len(parts) != 2 {
			return "некорректная длина команды"
		}
		if !student.backpack {
			return "некуда класть"
		}
		item := parts[1]
		if item == "дверь" || !student.position.items[item] {
			return "нет такого"
		}
		student.inventory = append(student.inventory, item)
		student.position.items[item] = false
		return fmt.Sprintf("предмет добавлен в инвентарь: %s", item)
	}
	return "неизвестная команда"
}
