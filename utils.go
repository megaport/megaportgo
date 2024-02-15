package megaport

import (
	"math/rand"
	"regexp"
	"time"
)

func IsGuid(guid string) bool {
	guidRegex := regexp.MustCompile(`(?mi)^[0-9a-f]{8}-[0-9a-f]{4}-[0-5][0-9a-f]{3}-[089ab][0-9a-f]{3}-[0-9a-f]{12}$`)

	if guidRegex.FindIndex([]byte(guid)) == nil {
		return false
	} else {
		return true
	}
}

func IsEmail(emailAddress string) bool {
	emailRegex := regexp.MustCompile(`(?mi)^[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}$`)

	if emailRegex.FindIndex([]byte(emailAddress)) == nil {
		return false
	} else {
		return true
	}
}

func GetTime(timestamp int64) time.Time {
	return time.Unix(timestamp/1000, 0)
}

func GenerateRandomVLAN() int {
	// exclude reserved values 0 and 4095 as per 802.1q
	return GenerateRandomNumber(1, 4094)
}

// GenerateRandomNumber generates a random number between an upper and lower bound.
func GenerateRandomNumber(lowerBound int, upperBound int) int {
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	return random.Intn(upperBound) + lowerBound
}
