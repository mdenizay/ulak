#!/usr/bin/env bash
# Ulak kurulum scripti
#
# Kullanım:
#   curl -fsSL https://mdenizay.github.io/ulak/install.sh | sudo bash

set -euo pipefail

RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; NC='\033[0m'
info()  { printf "${GREEN}[ulak]${NC} %s\n" "$*"; }
warn()  { printf "${YELLOW}[ulak]${NC} %s\n" "$*"; }
error() { printf "${RED}[ulak] HATA:${NC} %s\n" "$*" >&2; exit 1; }

# ── Root kontrolü ─────────────────────────────────────────────────────────
if [ "$(id -u)" -ne 0 ]; then
  error "Bu script root olarak çalışmalı. Deneyin: curl ... | sudo bash"
fi

# ── İşletim sistemi kontrolü ──────────────────────────────────────────────
[ -f /etc/os-release ] || error "İşletim sistemi tespit edilemedi. Yalnızca Debian/Ubuntu desteklenir."
# shellcheck source=/dev/null
. /etc/os-release
case "${ID:-}" in
  debian|ubuntu) ;;
  *) error "Desteklenmeyen OS: ${PRETTY_NAME:-bilinmiyor}. Debian/Ubuntu gerekli." ;;
esac

REPO_BASE="https://mdenizay.github.io/ulak"
KEYRING="/usr/share/keyrings/ulak-archive-keyring.gpg"
SOURCES="/etc/apt/sources.list.d/ulak.list"

# ── Gerekli araçlar ───────────────────────────────────────────────────────
info "Ön gereksinimler kuruluyor..."
apt-get update -qq
apt-get install -y -qq curl gpg

# ── GPG anahtarı ──────────────────────────────────────────────────────────
info "Ulak GPG anahtarı içe aktarılıyor..."
curl -fsSL "${REPO_BASE}/ulak-archive-keyring.gpg" \
  | gpg --dearmor \
  | tee "${KEYRING}" > /dev/null
chmod 644 "${KEYRING}"

# ── apt kaynağı ───────────────────────────────────────────────────────────
info "Ulak apt kaynağı ekleniyor..."
cat > "${SOURCES}" <<EOF
# Ulak apt deposu — https://github.com/mdenizay/ulak
deb [arch=amd64,arm64 signed-by=${KEYRING}] ${REPO_BASE} stable main
EOF
chmod 644 "${SOURCES}"

# ── Kurulum ───────────────────────────────────────────────────────────────
info "Paket listesi güncelleniyor..."
apt-get update -qq

info "Ulak kuruluyor..."
apt-get install -y ulak

# ── Doğrulama ─────────────────────────────────────────────────────────────
if command -v ulak > /dev/null 2>&1; then
  info "Başarıyla kuruldu: $(ulak version)"
else
  error "Kurulum tamamlandı fakat 'ulak' komutu bulunamadı."
fi

printf "\n${GREEN}Tamamlandı!${NC} Başlamak için: ulak --help\n"
printf "Güncellemek için: sudo apt upgrade ulak\n"
