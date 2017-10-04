fn main() {
    // println!("cargo:rustc-link-lib=static=rocksdb");
    println!("cargo:rustc-link-search=native=/home/jelte/go/src/github.com/GetStream/Keevo/.rocksdb-repo");
}
