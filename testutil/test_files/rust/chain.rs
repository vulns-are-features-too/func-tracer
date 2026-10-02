pub fn chain(_: bool) -> bool {
    chain1() && chain2() || chain3()
}

fn chain1() -> bool {
    true
}

fn chain2() -> bool {
    false
}

fn chain3() -> bool {
    true
}
