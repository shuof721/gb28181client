package media

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

// 串行化 ffmpeg，避免多通道同时抽流抢 CPU
var extractMu sync.Mutex

// NewMP4Source 从本地 MP4 抽出 Annex-B H.264 后循环播放。
func NewMP4Source(mp4Path string) (*FileSource, error) {
	if mp4Path == "" {
		return nil, fmt.Errorf("mp4_file is empty")
	}
	if _, err := os.Stat(mp4Path); err != nil {
		return nil, fmt.Errorf("mp4 not found: %s (%w)", mp4Path, err)
	}

	ff, err := findFFmpeg()
	if err != nil {
		return nil, err
	}

	cache := mp4CachePath(mp4Path)
	need, err := needExtract(mp4Path, cache)
	if err != nil {
		return nil, err
	}
	if need {
		extractMu.Lock()
		// double-check：并发时可能已被其他会话抽完
		need, err = needExtract(mp4Path, cache)
		if err == nil && need {
			log.Printf("[media] extracting H.264 from %s -> %s", mp4Path, cache)
			err = extractH264(ff, mp4Path, cache)
			if err == nil {
				if st, e := os.Stat(cache); e == nil {
					log.Printf("[media] extract done, %d bytes", st.Size())
				}
			}
		}
		extractMu.Unlock()
		if err != nil {
			return nil, err
		}
	}
	return NewFileSource(cache)
}

func mp4CachePath(mp4Path string) string {
	dir := filepath.Dir(mp4Path)
	base := strings.TrimSuffix(filepath.Base(mp4Path), filepath.Ext(mp4Path))
	// v3：限码率/缩放设置，废弃旧大文件缓存
	return filepath.Join(dir, base+".h264.v3.cache")
}

func needExtract(mp4, cache string) (bool, error) {
	stSrc, err := os.Stat(mp4)
	if err != nil {
		return false, err
	}
	stDst, err := os.Stat(cache)
	if err != nil {
		return true, nil
	}
	if stDst.Size() == 0 {
		return true, nil
	}
	// 源更新则重抽
	if stSrc.ModTime().After(stDst.ModTime()) {
		return true, nil
	}
	return false, nil
}

func findFFmpeg() (string, error) {
	// PATH 中查找
	if p, err := exec.LookPath("ffmpeg"); err == nil {
		return p, nil
	}
	// 常见 Windows 安装路径（本机曾检测到 gyan.dev 构建）
	candidates := []string{
		`C:\ffmpeg\bin\ffmpeg.exe`,
		`C:\Program Files\ffmpeg\bin\ffmpeg.exe`,
	}
	if home, err := os.UserHomeDir(); err == nil {
		candidates = append(candidates,
			filepath.Join(home, "scoop", "apps", "ffmpeg", "current", "bin", "ffmpeg.exe"),
			filepath.Join(home, "Downloads", "ffmpeg", "bin", "ffmpeg.exe"),
		)
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c, nil
		}
	}
	return "", fmt.Errorf("ffmpeg not found in PATH; install ffmpeg or set media.source to file/synthetic")
}

func extractH264(ffmpeg, mp4, out string) error {
	tmp := out + ".tmp"
	_ = os.Remove(tmp)

	// 国标实时预览：Baseline、无 B 帧、限制码率/分辨率，避免抽流过慢、缓存过大。
	args := []string{
		"-y", "-hide_banner", "-loglevel", "error",
		"-i", mp4,
		"-an", "-sn", "-dn",
		"-vf", "scale='min(1280,iw)':'min(720,ih)':force_original_aspect_ratio=decrease",
		"-c:v", "libx264",
		"-profile:v", "baseline",
		"-level", "3.1",
		"-preset", "ultrafast",
		"-tune", "zerolatency",
		"-pix_fmt", "yuv420p",
		"-bf", "0",
		"-g", "50",
		"-keyint_min", "25",
		"-sc_threshold", "0",
		"-b:v", "2500k",
		"-maxrate", "3000k",
		"-bufsize", "5000k",
		"-f", "h264",
		tmp,
	}
	cmd := exec.Command(ffmpeg, args...)
	if outb, err := cmd.CombinedOutput(); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("ffmpeg extract failed: %v\n%s", err, strings.TrimSpace(string(outb)))
	}

	st, err := os.Stat(tmp)
	if err != nil || st.Size() == 0 {
		_ = os.Remove(tmp)
		return fmt.Errorf("ffmpeg produced empty h264 from %s", mp4)
	}
	// 原子替换
	_ = os.Remove(out)
	if err := os.Rename(tmp, out); err != nil {
		// 跨卷时 fallback copy
		data, rerr := os.ReadFile(tmp)
		if rerr != nil {
			return rerr
		}
		if werr := os.WriteFile(out, data, 0o644); werr != nil {
			return werr
		}
		_ = os.Remove(tmp)
	}
	return nil
}
