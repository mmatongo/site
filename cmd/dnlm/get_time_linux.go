//go:build linux
// +build linux

package main

import (
	"syscall"
	"time"
	"unsafe"
)

func (a *app) getBirthTime(nativeInfo *syscall.Stat_t) time.Time {
	return getTimeFromTimespec(nativeInfo, unsafe.Pointer(&nativeInfo.Ctim))
}

func (a *app) getModTime(nativeInfo *syscall.Stat_t) time.Time {
	return getTimeFromTimespec(nativeInfo, unsafe.Pointer(&nativeInfo.Mtim))
}
