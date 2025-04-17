# --- s390x baseline z10 ---------------------------------
if [[ "${{ matrix.platform.target }}" == "s390x" ]]; then
    export CFLAGS="-march=z10 -mzarch"
    export CXXFLAGS="$CFLAGS"
    export RUSTFLAGS="-C target-cpu=z10"
fi
# --------------------------------------------------------

# If we're running on rhel centos, install needed packages.
if command -v yum &> /dev/null; then
    yum update -y && yum install -y perl-core openssl openssl-devel pkgconfig libatomic

    # If we're running on i686 we need to symlink libatomic
    # in order to build openssl with -latomic flag.
    if [[ ! -d "/usr/lib64" ]]; then
        ln -s /usr/lib/libatomic.so.1 /usr/lib/libatomic.so
    fi
else
    # If we're running on debian-based system.
    apt update -y && apt-get install -y libssl-dev openssl pkg-config
fi
