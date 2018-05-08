package accountmgr

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"
	"time"

	humanize "github.com/dustin/go-humanize"
	"github.com/fatih/color"
)

// List fetches and lists all accounts
func (am *AccountManager) List() error {

	files, err := ioutil.ReadDir(am.accountDir)
	if err != nil {
		return err
	}

	fmt.Println(fmt.Sprintf("%s%s%s",
		color.HiBlackString("Address"),
		strings.Repeat(" ", 32),
		color.HiBlackString("Date Created")),
	)

	for _, f := range files {
		if !f.IsDir() {
			m, _ := regexp.Match("[0-9]{10}_[a-zA-Z0-9]{34,}", []byte(f.Name()))
			if m {
				nameParts := strings.Split(f.Name(), "_")
				unixTime, _ := strconv.ParseInt(nameParts[0], 10, 64)
				timeCreated := time.Unix(unixTime, 0)
				fmt.Println(fmt.Sprintf("%s     %s", color.CyanString(nameParts[1]), humanize.Time(timeCreated)))
			}
		}
	}

	return nil
}
