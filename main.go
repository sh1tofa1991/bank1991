package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	ErrInsufficientFunds   = errors.New("недостаточно средств")
	ErrInvalidAmount       = errors.New("сумма должна быть положительной")
	ErrAccountNotFound     = errors.New("счёт не найден")
	ErrSameAccountTransfer = errors.New("нельзя перевести на тот же счёт")
)

type Transaction struct {
	Type      string
	Amount    float64
	Timestamp string
	Details   string
}

type Account struct {
	ID       string
	Owner    string
	Balance  float64
	History  []Transaction
}

func (a *Account) Deposit(amount float64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	a.Balance += amount
	a.History = append(a.History, Transaction{
		Type:      "пополнение",
		Amount:    amount,
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		Details:   "Пополнение счёта",
	})
	return nil
}

func (a *Account) Withdraw(amount float64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	if amount > a.Balance {
		return ErrInsufficientFunds
	}
	a.Balance -= amount
	a.History = append(a.History, Transaction{
		Type:      "снятие",
		Amount:    amount,
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		Details:   "Снятие средств",
	})
	return nil
}

func (a *Account) Transfer(to *Account, amount float64) error {
	if a.ID == to.ID {
		return ErrSameAccountTransfer
	}
	if amount <= 0 {
		return ErrInvalidAmount
	}
	if amount > a.Balance {
		return ErrInsufficientFunds
	}
	a.Balance -= amount
	to.Balance += amount
	now := time.Now().Format("2006-01-02 15:04:05")
	a.History = append(a.History, Transaction{
		Type:      "перевод",
		Amount:    amount,
		Timestamp: now,
		Details:   "Перевод на счёт " + to.ID,
	})
	to.History = append(to.History, Transaction{
		Type:      "перевод",
		Amount:    amount,
		Timestamp: now,
		Details:   "Перевод от счёта " + a.ID,
	})
	return nil
}

func (a *Account) GetBalance() float64 {
	return a.Balance
}

func (a *Account) GetStatement() string {
	result := "Выписка по счёту " + a.ID + " (" + a.Owner + "):\n"
	result += fmt.Sprintf("Баланс: %.2f\n", a.Balance)
	result += "История операций:\n"
	if len(a.History) == 0 {
		result += "  Нет операций.\n"
	} else {
		for _, tx := range a.History {
			result += fmt.Sprintf("  %s | %.2f | %s | %s\n",
				tx.Timestamp, tx.Amount, tx.Type, tx.Details)
		}
	}
	return result
}

type AccountService interface {
	Deposit(amount float64) error
	Withdraw(amount float64) error
	Transfer(to *Account, amount float64) error
	GetBalance() float64
	GetStatement() string
}

type Storage interface {
	SaveAccount(account *Account) error
	LoadAccount(accountID string) (*Account, error)
	GetAllAccounts() ([]*Account, error)
}

type InMemoryStorage struct {
	accounts map[string]*Account
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{accounts: make(map[string]*Account)}
}

func (s *InMemoryStorage) SaveAccount(account *Account) error {
	s.accounts[account.ID] = account
	return nil
}

func (s *InMemoryStorage) LoadAccount(accountID string) (*Account, error) {
	if acc, ok := s.accounts[accountID]; ok {
		return acc, nil
	}
	return nil, ErrAccountNotFound
}

func (s *InMemoryStorage) GetAllAccounts() ([]*Account, error) {
	accounts := make([]*Account, 0, len(s.accounts))
	for _, acc := range s.accounts {
		accounts = append(accounts, acc)
	}
	return accounts, nil
}

var storage = NewInMemoryStorage()

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	storage.SaveAccount(&Account{
		ID:      "ID0",
		Owner:   "Иван",
		Balance: 0,
		History: []Transaction{},
	})

	for {
		fmt.Println("\nМеню:")
		fmt.Println("1. Создать счёт")
		fmt.Println("2. Пополнить счёт")
		fmt.Println("3. Снять средства")
		fmt.Println("4. Перевести средства")
		fmt.Println("5. Показать баланс")
		fmt.Println("6. Показать выписку")
		fmt.Println("7. Список всех счетов")
		fmt.Println("0. Выйти")
		fmt.Print("Выберите пункт: ")

		scanner.Scan()
		choice := strings.TrimSpace(scanner.Text())

		switch choice {
		case "1":
			fmt.Print("Имя владельца: ")
			scanner.Scan()
			name := strings.TrimSpace(scanner.Text())
			if name == "" {
				fmt.Println("Имя не может быть пустым.")
				continue
			}
			id := "ID" + strconv.FormatInt(time.Now().UnixNano(), 10)
			acc := &Account{ID: id, Owner: name, Balance: 0}
			storage.SaveAccount(acc)
			fmt.Printf("Создан счёт: %s\n", acc.ID)

		case "2":
			fmt.Print("ID счёта: ")
			scanner.Scan()
			id := strings.TrimSpace(scanner.Text())
			acc, err := storage.LoadAccount(id)
			if err != nil {
				fmt.Println("Ошибка:", err)
				continue
			}
			fmt.Print("Сумма: ")
			scanner.Scan()
			amount, err := strconv.ParseFloat(scanner.Text(), 64)
			if err != nil {
				fmt.Println("Некорректная сумма")
				continue
			}
			err = acc.Deposit(amount)
			if err != nil {
				fmt.Println("Ошибка:", err)
			} else {
				fmt.Println("Пополнение успешно")
				storage.SaveAccount(acc)
			}

		case "3":
			fmt.Print("ID счёта: ")
			scanner.Scan()
			id := strings.TrimSpace(scanner.Text())
			acc, err := storage.LoadAccount(id)
			if err != nil {
				fmt.Println("Ошибка:", err)
				continue
			}
			fmt.Print("Сумма: ")
			scanner.Scan()
			amount, err := strconv.ParseFloat(scanner.Text(), 64)
			if err != nil {
				fmt.Println("Некорректная сумма")
				continue
			}
			err = acc.Withdraw(amount)
			if err != nil {
				fmt.Println("Ошибка:", err)
			} else {
				fmt.Println("Снятие успешно")
				storage.SaveAccount(acc)
			}

		case "4":
			fmt.Print("Ваш ID: ")
			scanner.Scan()
			fromID := strings.TrimSpace(scanner.Text())
			from, err := storage.LoadAccount(fromID)
			if err != nil {
				fmt.Println("Ошибка:", err)
				continue
			}
			fmt.Print("ID получателя: ")
			scanner.Scan()
			toID := strings.TrimSpace(scanner.Text())
			to, err := storage.LoadAccount(toID)
			if err != nil {
				fmt.Println("Ошибка:", ErrAccountNotFound)
				continue
			}
			fmt.Print("Сумма: ")
			scanner.Scan()
			amount, err := strconv.ParseFloat(scanner.Text(), 64)
			if err != nil {
				fmt.Println("Некорректная сумма")
				continue
			}
			err = from.Transfer(to, amount)
			if err != nil {
				fmt.Println("Ошибка:", err)
			} else {
				fmt.Println("Перевод выполнен")
				storage.SaveAccount(from)
				storage.SaveAccount(to)
			}

		case "5":
			fmt.Print("ID счёта: ")
			scanner.Scan()
			id := strings.TrimSpace(scanner.Text())
			acc, err := storage.LoadAccount(id)
			if err != nil {
				fmt.Println("Ошибка:", err)
				continue
			}
			fmt.Printf("Баланс: %.2f\n", acc.GetBalance())

		case "6":
			fmt.Print("ID счёта: ")
			scanner.Scan()
			id := strings.TrimSpace(scanner.Text())
			acc, err := storage.LoadAccount(id)
			if err != nil {
				fmt.Println("Ошибка:", err)
				continue
			}
			fmt.Println(acc.GetStatement())

		case "7":
			list, _ := storage.GetAllAccounts()
			if len(list) == 0 {
				fmt.Println("Нет счетов")
			} else {
				fmt.Println("Все счета:")
				for _, acc := range list {
					fmt.Printf("  %s | %s | %.2f\n", acc.ID, acc.Owner, acc.Balance)
				}
			}

		case "0":
			fmt.Println("Выход.")
			return

		default:
			fmt.Println("Неверный выбор.")
		}
	}
}