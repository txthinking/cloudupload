package cloudupload

import (
	"encoding/binary"
	"log"
	"strconv"
	"strings"
	"testing"

	uuid "github.com/satori/go.uuid"
)

func TestUUID(t *testing.T) {
	id := strings.Replace(uuid.NewV4().String(), "-", "", -1)
	log.Println(id)
	i := binary.BigEndian.Uint64([]byte(id))
	log.Println(i)
	s := strconv.FormatUint(i, 36)
	log.Println(s)
	s = ToShort(int64(i))
	log.Println(s)
}
