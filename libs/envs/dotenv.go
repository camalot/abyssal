package envs

import (
	"os"
	"path"
	"strings"

	"github.com/joho/godotenv"
)

func LoadDotEnv(files ...string) (error) {
  HOME, err := os.UserHomeDir()
  if err != nil {
    HOME = os.Getenv("HOME")
  }

  if HOME != "" {
    files = append(files, path.Join(HOME, ".env"))
  }
  files = append(files, "./.env")
  files = append(files, strings.ReplaceAll("~/.abyssal/.env", "~", HOME))
  // remove duplicates
  files = unique(files)

  for i, file := range files {
    files[i] = strings.ReplaceAll(file, "~", HOME)
  }
  err = godotenv.Load(files...)
  if err != nil {
		return err
  }
  return nil
}

func unique(files []string) []string {
  keys := make(map[string]bool)
  list := []string{}
  for _, entry := range files {
    if _, value := keys[entry]; !value {
      keys[entry] = true
      list = append(list, entry)
    }
  }
  return list
}
