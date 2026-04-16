# ToolsBank Go - Fix Planı

## 🔴 Öncelik 1: Bug Fixler

### BUG-001: TTL expires_at Hesaplama Hatası
**Dosya:** `pkg/mcp/server.go`
**Satır:** `memory_get` fonksiyonu
**Sorun:** `datetime('now')` SQLite'da timezone-aware değil, UTC karşılaştırması yapılmıyor
**Fix:** `datetime('now')` → `datetime('utcnow')` kullanılmalı

### BUG-002: Status Validation Eksik
**Dosya:** `pkg/mcp/server.go`
**Sorun:** `task_update` herhangi bir status değerini kabul ediyor
**Fix:** Geçerli status değerleri: pending, in_progress, completed, cancelled

### BUG-003: Task ID Formatı Zayıf
**Dosya:** `pkg/mcp/server.go`
**Sorun:** `fmt.Sprintf("task-%d", os.Getpid())` aynı PID'de çakışma riski
**Fix:** UUID kullanımı veya timestamp + random

---

## ⚠️ Öncelik 2: İyileştirmeler

### IMP-001: Semantic Search Optimizasyonu
**Dosya:** `pkg/mcp/server.go`
**Mevcut:** LIKE query (yavaş, basit)
**Hedef:** FTS5 (Full-Text Search) entegrasyonu
**Not:** Şimdilik basit FTS5, sonra gerçek vektör araması

### IMP-002: Input Validation
**Tüm tool handler'lar:** SQL injection koruması için parametrized queries zaten var
**Eksik:** Type validation, length limits

### IMP-003: Error Handling İyileştirmesi
**Dosya:** `pkg/utils/errors.go`
**Eksik:** Custom error types, error wrapping

---

## 📝 Öncelik 3: Test Coverage

### TEST-001: Unit Testler
- [x] Mevcut testler (server_test.go)
- [ ] `memory_get` TTL testi
- [ ] `task_update` status validation testi
- [ ] Edge case testleri

### TEST-002: Entegrasyon Testleri
- [ ] DB ile gerçek query testleri
- [ ] Transaction testleri
- [ ] Concurrent access testleri

### TEST-003: Manuel Test Scriptleri
- [ ] Test scripti (bash)
- [ ] Benchmark scripti

---

## 🚀 Öncelik 4: Build & Deployment

- [ ] Build system (Makefile)
- [ ] CI/CD pipeline
- [ ] Docker image

---

## 📊 Durum Takibi

| Bug | Durum | Agent |
|-----|-------|-------|
| BUG-001 | 🔄 Fixleniyor | DB-Agent |
| BUG-002 | 🔄 Fixleniyor | Tool-Agent |
| BUG-003 | 🔄 Fixleniyor | Tool-Agent |
| IMP-001 | 📋 Planlandı | DB-Agent |
| TEST-001 | 📋 Planlandı | Test-Agent |
| TEST-002 | 📋 Planlandı | Test-Agent |
| TEST-003 | 📋 Planlandı | Test-Agent |

---

## Agent Görev Dağılımı

| Agent | Görevler |
|-------|----------|
| DB-Agent | BUG-001, IMP-001 |
| Tool-Agent | BUG-002, BUG-003 |
| Test-Agent | TEST-001, TEST-002, TEST-003 |
| Security-Agent | Güvenlik review |
