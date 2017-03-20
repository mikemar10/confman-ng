package main

import (
    "fmt"
    "log"
    "os"
    "os/exec"
    "runtime"
    "time"
)

func main() {
    if runtime.GOOS != "linux" {
        log.Fatal("Linux is the only currently supported operating system")
    }
    squashfs_path := os.Args[1]
    tmp_mount_path := create_tmp_dir()
    mount_squashfs(squashfs_path, tmp_mount_path)
    cleanup_tmp_mount(tmp_mount_path)
}

func create_tmp_dir() string {
    tmpdir := fmt.Sprint(os.TempDir(), "/confman-ng-", time.Now().UnixNano())
    err := os.Mkdir(tmpdir, 0700)
    if err != nil {
        log.Fatal(err)
    }
    return tmpdir
}

func mount_squashfs(squashfs_path, mount_path string) {
    cmd := exec.Command("mount", squashfs_path, mount_path)
    if err := cmd.Run(); err != nil {
        log.Fatal(err)
    }
}

func cleanup_tmp_mount(mount_path string) {
    cmd := exec.Command("umount", mount_path)
    if err := cmd.Run(); err != nil {
        log.Fatal(err)
    }
    if err := os.Remove(mount_path); err != nil {
        log.Fatal(err)
    }
}
