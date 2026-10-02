use std::io::Write;

pub fn ext_hello() {
    _ = std::io::stdout().write(b"hello world");
}

pub fn ext_print(s: &str) {
    _ = std::io::stdout().write(s.as_bytes());
}
