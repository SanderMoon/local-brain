class LocalBrain < Formula
  desc "Minimalist, local-first project management system for terminal users"
  homepage "https://github.com/YOUR_USERNAME/local-brain"
  url "https://github.com/YOUR_USERNAME/local-brain/archive/v1.0.0.tar.gz"
  sha256 "REPLACE_WITH_SHA256_OF_YOUR_TARBALL"
  license "MIT"

  depends_on "jq"
  depends_on "ripgrep"
  depends_on "fzf"
  depends_on "bat"
  depends_on "syncthing"

  def install
    # Use the Makefile to install to Homebrew's prefix
    system "make", "install", "PREFIX=#{prefix}"
  end

  test do
    # Simple test to verify installation
    assert_match "Local Brain", shell_output("#{bin}/brain --help")
  end
end
