package libdeploy

import (
    "fmt"
    "log"
    "strings"
    "io/ioutil"
    "strconv"
    "bytes"
    "time"
    "github.com/BurntSushi/toml"
)

var timeFormat = "2006-01-02T15:04:05Z"

type Config map[string]interface{}

func ReadConfig(fileName string) Config {
    blob, err := ioutil.ReadFile(fileName)
    if err != nil {
        fmt.Println("Error!!!!")
        panic(err)
    }

    var config Config
    _, err = toml.Decode(string(blob), &config)
    if err != nil {
        fmt.Println("Error parsing toml!")
        panic(err)
    }

    return config
}

func PrintConfig(config Config) {
    buf := new(bytes.Buffer)
    if err := toml.NewEncoder(buf).Encode(config); err != nil {
        log.Fatal("Cannot print config! ", err)
    }

    fmt.Println(buf.String())
}

func PutVariable(path string, config Config) {
    var full_path = "config"
    var buffer, changer map[string]interface{}
    var last_path string

    log.Printf("setting %s\n", path)
    buf := strings.Split(path, ":")
    // FIXME add check that only one semicolon is used
    path = buf[0]
    val := strings.Join(buf[1:], ":")
    variable_path := strings.Split(path,".")
    for num, p := range variable_path {
        full_path = fmt.Sprintf("%s[%s]", full_path, p)
        if buffer == nil {
            if num + 1 < len(variable_path) { // first element
                if config[p] == nil { // if no middle element
                    config[p] = map[string]interface{}{}
                }
                buffer = config[p].(map[string]interface{})
            } else { // first and last
                changer = config
                last_path = p
            }
        } else {
            if num + 1 < len(variable_path) { // middle element
                if buffer[p] == nil { // if no middle element
                    buffer[p] = map[string]interface{}{}
                }
                buffer = buffer[p].(map[string]interface{})
                fmt.Println(buffer)
            } else { // last element
                changer = buffer
                last_path = p
            }
        }
    }

    if changer[last_path] != nil {
        switch t := changer[last_path].(type) {
        case time.Time:
            t, err := time.Parse(timeFormat, val)
            if err != nil {
                log.Fatal("You shold specify time in format %s", timeFormat)
            }

            changer[last_path] = t
        case int:
            i, err := strconv.Atoi(val)
            if err != nil {
                log.Fatal(fmt.Sprintf("%s should be int!", path))
            }

            changer[last_path] = i
        case string:
            changer[last_path] = val
        default:
            log.Fatal(fmt.Sprintf("To set %s, value should be %T!", path, t))
        }
    } else {
        if t, err := time.Parse(timeFormat, val); err != nil {
            if i, err := strconv.Atoi(val); err != nil {
                if r, err := strconv.ParseFloat(val, 64); err != nil {
                    changer[last_path] = val // Cannot conver to any type, sujesting string
                } else { // Converted to float
                    changer[last_path] = r
                }
            } else { // Converted to int
                changer[last_path] = i
            }
        } else { // Converted to time
            changer[last_path] = t
        }
    }
}
