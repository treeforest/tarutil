package tarutil

import (
	"archive/tar"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/gookit/goutil/fsutil"
	"github.com/klauspost/compress/gzip"
)

const (
	suffixTar   = ".tar"
	suffixTarGz = ".tar.gz"
)

// Extract 解压缩归档文件。将压缩文件（src）解压到目标文件夹（dst）。
// 参数：
//   - src: 待解压的压缩文件，压缩文件应以 .tar 或 .tar.gz 为文件后缀。
//   - dst: 存放解压后文件的文件夹路径，若为空，则默认解压到当前文件夹。
func Extract(src, dst string) error {
	if !strings.HasSuffix(src, suffixTar) && !strings.HasSuffix(src, suffixTarGz) {
		return fmt.Errorf("src should hava .tar or .tar.gz extension")
	}

	if dst == "" {
		dst = "."
	}
	absPath, err := filepath.Abs(dst)
	if err != nil {
		return fmt.Errorf("failed to filepath.Abs: %s", dst)
	}
	dst = filepath.Clean(absPath)

	err = os.MkdirAll(dst, 0755)
	if err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	archiveFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open src: %w", err)
	}
	defer archiveFile.Close()

	var r io.Reader = archiveFile

	if strings.HasSuffix(src, suffixTarGz) {
		gzipReader, err := gzip.NewReader(archiveFile)
		if err != nil {
			return fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzipReader.Close()
		r = gzipReader
	}

	tarReader := tar.NewReader(r)
	for {
		header, err := tarReader.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return fmt.Errorf("failed to read tar header: %w", err)
		}

		target := filepath.Clean(filepath.Join(dst, header.Name)) //nolint:gosec
		if !strings.HasPrefix(target, dst) {
			return fmt.Errorf("target '%s' does not have the expected prefix: %s", target, dst)
		}

		if header.FileInfo().IsDir() {
			if err = os.MkdirAll(target, header.FileInfo().Mode()); err != nil {
				return fmt.Errorf("failed to create directory '%s': %w", target, err)
			}
		} else {
			targetFile, err := os.OpenFile(target, os.O_RDWR|os.O_CREATE|os.O_TRUNC, header.FileInfo().Mode())
			if err != nil {
				return fmt.Errorf("failed to create target file: %w", err)
			}
			if _, err = io.Copy(targetFile, tarReader); err != nil { //nolint:gosec
				_ = targetFile.Close()
				return fmt.Errorf("failed to extract file: %w", err)
			}
			_ = targetFile.Close()
		}
	}

	return nil
}

// Archive 压缩文件。将待压缩的文件或目录（src）压缩到目标压缩文件（dst）中。
// 参数：
//   - src：待压缩的文件或目录
//   - dst：目标压缩文件，应以 .tar 或 .tar.gz 结尾
//   - excludePaths: 在压缩过程中需要跳过的文件或目录路径列表
func Archive(src, dst string, excludePaths ...string) error {
	if !pathExists(src) {
		return fmt.Errorf("path %s not exists", src)
	}

	if !pathExists(filepath.Dir(dst)) {
		err := os.MkdirAll(filepath.Dir(dst), 0755)
		if err != nil {
			return fmt.Errorf("failed to create directory '%s': %w", filepath.Dir(dst), err)
		}
	}

	if !strings.HasSuffix(dst, suffixTar) && !strings.HasSuffix(dst, suffixTarGz) {
		return fmt.Errorf("dst should hava .tar or .tar.gz extension")
	}

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	var tw *tar.Writer
	if strings.HasSuffix(dst, suffixTarGz) {
		// create gzip.Writer
		gw := gzip.NewWriter(dstFile)
		defer gw.Close()
		// create tar.Writer
		tw = tar.NewWriter(gw)
		defer tw.Close()
	} else {
		// create tar.Writer
		tw = tar.NewWriter(dstFile)
		defer tw.Close()
	}

	// 开始压缩
	if fsutil.IsDir(src) {
		// ----> 压缩目录
		// 遍历文件夹并将文件添加到tar.Writer中
		err = filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// 跳过 .
			if src == path {
				return nil
			}

			// 跳过文件
			for _, excludePath := range excludePaths {
				if strings.Contains(path, excludePath) {
					return nil
				}
			}

			// 创建tar.Header
			header, err := tar.FileInfoHeader(info, info.Name())
			if err != nil {
				return err
			}

			// 更新文件路径
			relPath, err := filepath.Rel(src, path)
			if err != nil {
				return err
			}
			header.Name = relPath

			// 写入tar.Header
			err = tw.WriteHeader(header)
			if err != nil {
				return err
			}

			// 如果是文件，将文件内容写入tar.Writer
			if !info.IsDir() {
				file, err := os.Open(path)
				if err != nil {
					return err
				}
				defer file.Close()

				_, err = io.Copy(tw, file)
				if err != nil {
					return err
				}
			}

			return nil
		})
	} else {
		// ----> 压缩文件
		// 打开源文件
		srcFile, err := os.Open(src)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		// 获取源文件信息
		info, err := srcFile.Stat()
		if err != nil {
			return err
		}

		// 创建tar.Header
		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return err
		}

		// 写入tar.Header
		err = tw.WriteHeader(header)
		if err != nil {
			return err
		}

		// 将源文件内容复制到tar.Writer
		_, err = io.Copy(tw, srcFile)
		if err != nil {
			return err
		}

		return nil
	}

	return err
}

func pathExists(path string) bool {
	if path == "" {
		return false
	}

	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
