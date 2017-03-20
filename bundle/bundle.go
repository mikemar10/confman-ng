package main

import (
    "crypto/sha256"
    "fmt"
    "github.com/davecheney/xattr"
    "io"
    "io/ioutil"
    "log"
    "os"
    "os/exec"
    "os/user"
    "path/filepath"
    "strconv"
    "strings"
    "syscall"
)

func write_map(path string, items map[string]uint32) error {
    data := make([]byte, 0)
    for key, value := range items {
        s := fmt.Sprintf("%s:%d\n", key, value)
        data = append(data, s...)
    }

    ioutil.WriteFile(path, data, 0444)
    return nil
}

func main() {
    target := strings.TrimRight(os.Args[1], "/")
    all_users := make(map[string]uint32)
    all_groups := make(map[string]uint32)
    fmt.Printf("Adding metadata...\n")

    preserve_metadata := func(path string, info os.FileInfo, err error) error {
        if err != nil {
            fmt.Println(err)
            return nil
        }

        stat, ok := info.Sys().(*syscall.Stat_t)
        if !ok {
            fmt.Printf("Not a syscall.Stat_t???\n")
            return nil
        }
        username, _ := user.LookupId(strconv.FormatUint(uint64(stat.Uid), 10))
        groupname, _ := user.LookupGroupId(strconv.FormatUint(uint64(stat.Gid), 10))
        fmt.Printf("%s:%s %s\n", username.Username, groupname.Name, path)

        switch mode := info.Mode(); {
            case mode.IsDir():
                xattr.Setxattr(path, "user.user", []byte(username.Username))
                xattr.Setxattr(path, "user.group", []byte(groupname.Name))
            case mode.IsRegular():
                xattr.Setxattr(path, "user.user", []byte(username.Username))
                xattr.Setxattr(path, "user.group", []byte(groupname.Name))
                f, _ := os.Open(path)
                h := sha256.New()
                io.Copy(h, f)
                f.Close()
                fmt.Printf("%x\n", h.Sum(nil))
                sum := []byte(fmt.Sprintf("%x", h.Sum(nil)))
                xattr.Setxattr(path, "user.sha256sum", sum)
        }
        all_users[username.Username] = stat.Uid
        all_groups[groupname.Name] = stat.Gid
        return nil
    }

    filepath.Walk(target, preserve_metadata)
    write_map(target + "/.users", all_users)
    write_map(target + "/.groups", all_groups)
    cmd := exec.Command("mksquashfs", target, "target.squashfs", "-no-progress")
    stdoutStderr, err := cmd.CombinedOutput()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("%s\n", stdoutStderr)
}
