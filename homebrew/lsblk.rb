# lsblk.rb
class Lsblk < Formula
  desc "Description of your lsblk tool"
  homepage "https://github.com/JakeTrock/gosblk"
  url "https://github.com/JakeTrock/gosblk/archive/refs/tags/v1.tar.gz"
  sha256 "be6116ae1c07cd0b794b189a537aae302c52aa0257e2211551c8e6d2bd8a8a55"
  license "GNU GPLv3"

  depends_on "go" => :build

  def install
    system "make", "build"
    bin.install "bin/lsblk"
  end

  test do
    system "#{bin}/lsblk", "-h"
  end
end