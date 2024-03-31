# Maintainer: MotorTruck1221 motortruck1221@protonmail.com

pkgname=wisp-server-go
pkgver=0.0.1
pkgrel=1
pkgdesc="A Wisp server written in Go."
url="https://github.com/motortruck1221/wisp-go"
arch=('x86_64')
license=('AGPL3')
# the source is an artifact from the CI
source=("https://github.com/motortruck1221/wisp-go/raw/main/bin/wisp-server-go")
sha256sums=('SKIP')
package() {
    install -d "${pkgdir}/usr/bin"
    install -Dm755 "${srcdir}/wisp-server-go" "${pkgdir}/usr/bin/wisp-server-go"
}
