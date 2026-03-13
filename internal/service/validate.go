package service

func validateLogin(login string) error {
	if len(login) == 0 {
		return ErrEmptyLogin
	}
	return nil
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return ErrPasswordTooShort
	}
	return nil
}

func validateOrder(orderNumber string) error {
	if !checkLuhn(orderNumber) {
		return ErrInvalidOrderNumber
	}

	return nil
}

func checkLuhn(number string) bool {
	n := len(number)
	if n == 0 {
		return false
	}

	var sum int
	parity := n % 2
	for i := 0; i < n; i++ {
		if number[i] < '0' || number[i] > '9' {
			return false
		}
		digit := int(number[i] - '0')
		if i%2 == parity {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
	}
	return sum%10 == 0
}
