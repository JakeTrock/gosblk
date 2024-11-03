# lsblk.rb
class lsblk < Formula
  desc "LSBLK for macos implemented in golang"
  homepage "https://github.com/JakeTrock/gosblk"
  url "https://github.com/JakeTrock/gosblk/archive/refs/tags/v1.0.0.tar.gz"
  sha256 "fcd9bc61499ea0b02736a22f6b7da016fe6ca838793c9636bce3efdb8418b257"
  license "GNU GPLv3"

  def install
    bin.install "bin/lsblk"
  end

  test do
    system "#{bin}/lsblk", "-h"
  end
end