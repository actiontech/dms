package io

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/user"
	"strconv"
	"strings"
	"time"

	"github.com/actiontech/dms/pkg/dms-common/pkg/log"
)

func ReadFile(logger log.Logger, filePath string) (bs []byte, err error) {
	log.NewHelper(logger).Debugf("[ReadFile.start] path: %v", filePath)
	defer func() {
		log.NewHelper(logger).Debugf("[ReadFile.end] path: %v err: %v", filePath, err)
	}()
	return os.ReadFile(filePath)
}

func WriteFile(logger log.Logger, path string, content string, owner string, perm os.FileMode) (err error) {
	log.NewHelper(logger).Debugf("[WriteFile.start] path: %v, owner: %v, perm: %v", path, owner, perm)
	defer func() {
		log.NewHelper(logger).Debugf("[WriteFile.end] path: %v, owner: %v, perm: %v, err: %v", path, owner, perm, err)
	}()

	if err := EnsureFile(logger, path, owner, perm); nil != err {
		return err
	}
	return os.WriteFile(path, []byte(content), perm)
}

func WriteFileByReader(logger log.Logger, path string, reader io.Reader, owner string, perm os.FileMode) (err error) {
	log.NewHelper(logger).Debugf("[WriteFileByReader.start] path: %v, owner: %v, perm: %v", path, owner, perm)
	defer func() {
		log.NewHelper(logger).Debugf("[WriteFileByReader.end] path: %v, owner: %v, perm: %v, err: %v", path, owner, perm, err)
	}()

	if err := EnsureFile(logger, path, owner, perm); nil != err {
		return err
	}

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC, perm)
	if nil != err {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, reader)
	return err
}

func EnsureFile(logger log.Logger, path, owner string, perm os.FileMode) (err error) {
	log.NewHelper(logger).Debugf("[EnsureFile.start] path: %v, owner: %v, perm: %v", path, owner, perm)
	defer func() {
		log.NewHelper(logger).Debugf("[EnsureFile.end] path: %v, owner: %v, perm: %v, err: %v", path, owner, perm, err)
	}()

	uid, gids, err := LookupUidGidByUser(owner)
	if nil != err {
		return err
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, perm)
	if nil != err {
		return err
	}
	if err := f.Chmod(perm); nil != err {
		return err
	}
	if err := f.Chown(uid, gids[0]); nil != err {
		return err
	}
	if err := f.Close(); nil != err {
		return err
	}
	return nil
}

func IsFileExist(path string) bool {
	_, err := os.Stat(path)
	return nil == err
}

func Remove(logger log.Logger, path string) (err error) {
	log.NewHelper(logger).Debugf("[Remove.start] path: %v", path)
	defer func() {
		log.NewHelper(logger).Debugf("[Remove.end] path: %v, err: %v", path, err)
	}()

	if "" == strings.TrimRight(path, "/* ") {
		panic("cannot rm / path")
	}
	return os.RemoveAll(path)
}

func LookupUidGidByUser(username string) (uid int, gids []int, err error) {
	if "" != username {
		u, err := user.Lookup(username)
		if nil != err {
			return 0, []int{}, err
		}
		uid, _ = strconv.Atoi(u.Uid)

		// On different systems, gid's sorting is not the same.
		// In order to get a consistent gid, we set the primary gid in the first place.
		pGid, _ := strconv.Atoi(u.Gid)
		gids = []int{pGid}
		{
			groupIds, err := u.GroupIds()
			if nil != err {
				return 0, []int{}, err
			}
			for _, gidStr := range groupIds {
				gid, _ := strconv.Atoi(gidStr)
				if pGid == gid {
					continue
				}
				gids = append(gids, gid)
			}
		}
	} else {
		uid = os.Getuid()
		pGid := os.Getgid()
		gids = []int{pGid}
		if groups, err := os.Getgroups(); nil == err {
			for _, gid := range groups {
				if pGid == gid {
					continue
				}
				gids = append(gids, gid)
			}

		}
	}

	return uid, gids, nil
}

// 大部分组件都用到了ErrTimeout，变更影响较大，因此忽略该lint
type ErrTimeout string //nolint:errname

func (e ErrTimeout) Error() string {
	return "Err: time out: " + string(e)
}

// CopyWithTimeout is an Enhancement of io.Copy in golang std lib
func CopyWithTimeout(dst io.Writer, src io.Reader, timeoutSeconds int) (written int64, err error) {
	return copyBufferWithTimeout(dst, src, nil, timeoutSeconds)
}

// Enhancement of io.copyBuffer in golang std lib
func copyBufferWithTimeout(dst io.Writer, src io.Reader, buf []byte, timeoutSeconds int) (written int64, err error) {
	// If the reader has a WriteTo method, use it to do the copy.
	// Avoids an allocation and a copy.
	if wt, ok := src.(io.WriterTo); ok {
		return wt.WriteTo(dst)
	}
	// Similarly, if the writer has a ReadFrom method, use it to do the copy.
	if rt, ok := dst.(io.ReaderFrom); ok {
		return rt.ReadFrom(src)
	}
	if buf == nil {
		size := 32 * 1024
		if l, ok := src.(*io.LimitedReader); ok && int64(size) > l.N {
			if l.N < 1 {
				size = 1
			} else {
				size = int(l.N)
			}
		}
		buf = make([]byte, size)
	}
	begin := time.Now()
	timeoutChan := make(chan struct{}, 1)
	endFromTimeout := begin.Add(time.Second * time.Duration(timeoutSeconds))
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if nw < 0 || nr < nw {
				nw = 0
				if ew == nil {
					ew = errors.New("invalid write result")
				}
			}
			written += int64(nw)
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}

		if time.Now().After(endFromTimeout) {
			timeoutChan <- struct{}{}
		}

		select {
		case <-timeoutChan:
			err = ErrTimeout(fmt.Sprintf("copy buffer time out, limit [%v] seconds", timeoutSeconds))
			return 0, err
		default:
		}
	}
	return written, err
}
