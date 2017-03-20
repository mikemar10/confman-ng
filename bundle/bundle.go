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

func main() {
    target_path := strings.TrimRight(os.Args[1], "/")
    tag_files(target_path)
    make_squashfs(target_path)
}

func tag_files(target_path string) {
    all_users := make(map[string]uint32)
    all_groups := make(map[string]uint32)

    filepath.Walk(target_path, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            log.Fatal(err)
        }

        stat, ok := info.Sys().(*syscall.Stat_t)
        if !ok {
            log.Fatal("Failed type assertion of type syscall.Stat_t")
        }
        username := get_username(stat.Uid)
        groupname := get_groupname(stat.Gid)

        switch mode := info.Mode(); {
            case mode.IsRegular():
                xattr.Setxattr(path, "user.sha256sum", sha256sum(path))
                fallthrough
            case mode.IsDir():
                xattr.Setxattr(path, "user.user", []byte(username))
                xattr.Setxattr(path, "user.group", []byte(groupname))
        }
        all_users[username] = stat.Uid
        all_groups[groupname] = stat.Gid
        return nil
    })

    persist_map(target_path + "/.users", all_users)
    persist_map(target_path + "/.groups", all_groups)
}

func get_username(_uid uint32) string {
    uid := strconv.FormatUint(uint64(_uid), 10)
    username, err := user.LookupId(uid)
    if err != nil {
        log.Fatal(err)
    }
    return username.Username
}

func get_groupname(_gid uint32) string {
    gid := strconv.FormatUint(uint64(_gid), 10)
    groupname, err := user.LookupGroupId(gid)
    if err != nil {
        log.Fatal(err)
    }
    return groupname.Name
}

func sha256sum(path string) []byte {
    file, err := os.Open(path);
    if err != nil {
        log.Fatal(err)
    }
    hash := sha256.New()
    if _, err := io.Copy(hash, file); err != nil {
        log.Fatal(err)
    }
    if err := file.Close(); err != nil {
        log.Fatal(err)
    }
    return []byte(fmt.Sprintf("%x", hash.Sum(nil)))
}

func persist_map(path string, items map[string]uint32) error {
    data := make([]byte, 0)
    for key, value := range items {
        s := fmt.Sprintf("%s:%d\n", key, value)
        data = append(data, s...)
    }

    ioutil.WriteFile(path, data, 0444)
    return nil
}

func make_squashfs(target_path string) {
    cmd := exec.Command("mksquashfs", target_path, "target.squashfs", "-no-progress")
    output, err := cmd.CombinedOutput()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("%s\n", output)
}
