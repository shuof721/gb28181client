package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/local/gb28181-device/internal/config"
)

// Store 基于目录的 JSON 设备配置存储器。
type Store struct {
	dir string
	mu  sync.RWMutex
}

func New(dir string) *Store {
	return &Store{dir: dir}
}

// Dir 返回存储根目录。
func (s *Store) Dir() string {
	return s.dir
}

// EnsureDir 创建存储目录（若不存在）。
func (s *Store) EnsureDir() error {
	return os.MkdirAll(s.dir, 0o755)
}

// List 列出所有设备配置，按 ID 升序排序。
func (s *Store) List() ([]*config.DeviceProfile, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries, err := os.ReadDir(s.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*config.DeviceProfile{}, nil
		}
		return nil, fmt.Errorf("read store dir: %w", err)
	}

	var profiles []*config.DeviceProfile
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".json") {
			continue
		}
		p, err := s.readLocked(filepath.Join(s.dir, e.Name()))
		if err != nil {
			// 跳过异常文件
			continue
		}
		profiles = append(profiles, p)
	}

	sort.Slice(profiles, func(i, j int) bool {
		return profiles[i].Device.ID < profiles[j].Device.ID
	})
	return profiles, nil
}

// Get 获取指定国标 ID 的设备配置。
func (s *Store) Get(id string) (*config.DeviceProfile, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	id = config.NormalizeGBID(id)
	path := filepath.Join(s.dir, id+".json")
	return s.readLocked(path)
}

// Save 保存或更新设备配置（采用原子临时文件替换）。
func (s *Store) Save(p *config.DeviceProfile) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.EnsureDir(); err != nil {
		return err
	}

	p.ApplyDefaults()
	if err := p.Validate(); err != nil {
		return fmt.Errorf("validate device profile: %w", err)
	}

	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal device profile: %w", err)
	}

	id := p.Device.ID
	targetPath := filepath.Join(s.dir, id+".json")
	tmpPath := filepath.Join(s.dir, id+".json.tmp")

	if err := os.WriteFile(tmpPath, data, 0o644); err != nil {
		return fmt.Errorf("write tmp file: %w", err)
	}

	// Windows 下直接 rename 若目标已存在可能会报错，先删除目标确保幂等
	_ = os.Remove(targetPath)
	if err := os.Rename(tmpPath, targetPath); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("atomic rename failed: %w", err)
	}

	return nil
}

// Delete 删除指定国标 ID 的设备文件。
func (s *Store) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	id = config.NormalizeGBID(id)
	path := filepath.Join(s.dir, id+".json")
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return os.Remove(path)
}

// ImportLegacyYAML 自动从旧版 YAML 配置迁移初始设备（若数据目录为空）。
func (s *Store) ImportLegacyYAML(yamlPath string) (*config.DeviceProfile, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.EnsureDir(); err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(s.dir)
	if err == nil {
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(strings.ToLower(e.Name()), ".json") {
				// 已存在设备数据，不重复导入
				return nil, nil
			}
		}
	}

	if _, err := os.Stat(yamlPath); err != nil {
		return nil, nil
	}

	cfg, err := config.Load(yamlPath)
	if err != nil {
		return nil, fmt.Errorf("load legacy yaml: %w", err)
	}

	profile := &config.DeviceProfile{
		Enabled: true,
		SIP:     cfg.SIP,
		Device:  cfg.Device,
		Media:   cfg.Media,
	}
	profile.ApplyDefaults()

	data, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		return nil, err
	}

	targetPath := filepath.Join(s.dir, profile.Device.ID+".json")
	if err := os.WriteFile(targetPath, data, 0o644); err != nil {
		return nil, err
	}

	return profile, nil
}

func (s *Store) readLocked(path string) (*config.DeviceProfile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var p config.DeviceProfile
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("unmarshal profile %s: %w", filepath.Base(path), err)
	}
	p.ApplyDefaults()
	return &p, nil
}
