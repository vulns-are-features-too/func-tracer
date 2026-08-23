pub fn recurse_self(level: i8) {
    if level > 0 {
        recurse_self(level - 2)
    }
}

pub fn recurse_other(num: u8) {
    recurse_or_return(num - 1)
}

pub fn recurse_or_return(num: u8) {
    if num == 0 {
        return;
    }
    recurse_other(num - 1)
}
