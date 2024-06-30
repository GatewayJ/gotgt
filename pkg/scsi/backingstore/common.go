/*
Copyright 2017 The GoStor Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package backingstore

import (
	"fmt"
	"io"
	"os"
	"strconv"

	log "github.com/sirupsen/logrus"

	"github.com/gostor/gotgt/pkg/api"
	"github.com/gostor/gotgt/pkg/scsi"
	"github.com/gostor/gotgt/pkg/util"
)

const (
	FileBackingStorage = "file"
)

func init() {
	scsi.RegisterBackingStore(FileBackingStorage, new)
}

type FileBackingStore struct {
	scsi.BaseBackingStore
	file *os.File
}

func new() (api.BackingStore, error) {
	return &FileBackingStore{
		BaseBackingStore: scsi.BaseBackingStore{
			Name:            FileBackingStorage,
			DataSize:        0,
			OflagsSupported: 0,
		},
	}, nil
}

func (bs *FileBackingStore) Open(dev *api.SCSILu, path string) error {
	var mode os.FileMode

	finfo, err := os.Stat(path)
	if err != nil {
		return err
	} else {
		// determine file type
		mode = finfo.Mode()
	}

	f, err := os.OpenFile(path, os.O_RDWR, os.ModePerm)

	if err == nil {
		// block device filesize needs to be treated differently
		if (mode & os.ModeDevice) != 0 {
			pos, err := f.Seek(0, io.SeekEnd)
			if err != nil {
				return err
			}
			bs.DataSize = uint64(pos)
		} else {
			if finfo == nil {
				log.Infof("finfo is nil")
			}
			bs.DataSize = uint64(finfo.Size())
		}
	}

	bs.file = f
	return err
}

func (bs *FileBackingStore) Close(dev *api.SCSILu) error {
	return bs.file.Close()
}

func (bs *FileBackingStore) Init(dev *api.SCSILu, Opts string) error {
	return nil
}

func (bs *FileBackingStore) Exit(dev *api.SCSILu) error {
	return nil
}

func (bs *FileBackingStore) Size(dev *api.SCSILu) uint64 {
	return bs.DataSize
}

var s3FilePath = "/var/tmp/file_"

func (bs *FileBackingStore) Read(offset, tl int64) ([]byte, error) {
	log.Info("read")
	fileNum := offset / 1024 / 1024 / 1024      // 文件编号 从0开始
	fileOffset := offset % (1024 * 1024 * 1024) // 文件偏移量 开始待读取位置

	tmpbuf := make([]byte, 0)
	for i := 0; i < int(tl); i++ { //todo 这样读的很慢，目录的元数据受损。如果不跨文件读取是正常的
		var tmp = make([]byte, 1)
		filePath := s3FilePath + strconv.FormatInt(fileNum, 10)
		f, err := os.OpenFile(filePath, os.O_RDWR, 0666)
		if err != nil {
			log.Error(err)
			return nil, err
		}

		n, err := f.ReadAt(tmp, fileOffset)
		log.Infof("read length %d, offset %d, fileNum %d", n, fileOffset, fileNum)
		if err == io.EOF {
			log.Errorf("read EOF")
		} else if err != nil {
			log.Error(err)
			return nil, err
		}
		fileOffset++
		offset++
		tmpbuf = append(tmpbuf, tmp...)
		f.Close()
	}

	if tl != int64(len(tmpbuf)) {
		log.Error("read != ")
		return nil, fmt.Errorf("read is not same length of length")
	}
	return tmpbuf, nil
}

func (bs *FileBackingStore) Write(wbuf []byte, offset int64) error {

	fileNum := offset / (1024 * 1024 * 1024)                                // 文件编号 从0开始
	fileNumComplete := (offset + int64(len(wbuf)) - 1) / 1024 / 1024 / 1024 //  写完的文件编号
	fileOffset := offset % (1024 * 1024 * 1024)                             // 文件偏移量 开始待写入位置
	log.Infof("write filenum %d", fileNum)
	filePath := s3FilePath + strconv.FormatInt(fileNum, 10)
	f, err := os.OpenFile(filePath, os.O_RDWR, 0666)
	defer f.Close()
	if err != nil {
		log.Error(err)
		return err
	}
	if fileNum == fileNumComplete {
		// 没有跨越文件
		_, err := f.WriteAt(wbuf, offset)
		// log.Infof("write length %d, offset %d", length, offset)
		if err != nil {
			log.Error(err)
			return err
		}
	} else {
		// 跨越文件
		length, err := f.WriteAt(wbuf[:1024*1024*1024-fileOffset], fileOffset)
		log.Infof("write length %d, offset %d", length, offset)
		if err != nil {
			log.Error(err)
			return err
		}
		length, err = f.WriteAt(wbuf[1024*1024*1024-fileOffset:], 0)
		log.Infof("write length %d, offset %d", length, offset)
		if err != nil {
			log.Error(err)
			return err
		}
	}
	f.Sync()
	// if length != len(wbuf) {
	// 	return fmt.Errorf("write is not same length of length")
	// }
	return nil
}

func (bs *FileBackingStore) DataSync(offset, tl int64) error {
	log.Info("sync")
	for i := 0; i < 11; i++ {

		filePath := s3FilePath + strconv.FormatInt(int64(i), 10)
		f, err := os.OpenFile(filePath, os.O_RDWR, 0666)
		if err != nil {
			log.Error(err)
			return err
		}
		// f.Sync()
		util.Fdatasync(f)
		f.Close()
	}

	// return util.Fdatasync(bs.file)
	return nil
}

func (bs *FileBackingStore) DataAdvise(offset, length int64, advise uint32) error {
	return util.Fadvise(bs.file, offset, length, advise)
}

func (bs *FileBackingStore) Unmap([]api.UnmapBlockDescriptor) error {
	return nil
}
