//go:build darwin
// +build darwin

package main

import (
	"syscall"
	"time"
	"unsafe"
)

func (a *app) getBirthTime(nativeInfo *syscall.Stat_t) time.Time {
	return getTimeFromTimespec(nativeInfo, unsafe.Pointer(&nativeInfo.Ctimespec))
}

func (a *app) getModTime(nativeInfo *syscall.Stat_t) time.Time {
	return getTimeFromTimespec(nativeInfo, unsafe.Pointer(&nativeInfo.Mtimespec))
}
