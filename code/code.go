package code

import (
	"fmt"
)

func BuildFRN(domain string, entity string, shortCode string) string {
	return fmt.Sprintf("frn:%s:%s:%s", domain, entity, shortCode)
}
