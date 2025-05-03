package entity

import "fmt"

/*TODO replace on room:<roomID> channel*/
func GroupChannel(groupID uint) string {
	return fmt.Sprintf("group:%d", groupID)
}
