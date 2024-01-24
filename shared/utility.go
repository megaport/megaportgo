// Copyright 2020 Megaport Pty Ltd
//
// Licensed under the Mozilla Public License, Version 2.0 (the
// "License"); you may not use this file except in compliance with
// the License. You may obtain a copy of the License at
//
//       https://mozilla.org/MPL/2.0/
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// The `shared` package is houses functions that are used throughout the entire Megaport Go Library. They are not meant
// for general use outside of the Library.
package shared

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

func GetCurrentTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
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
