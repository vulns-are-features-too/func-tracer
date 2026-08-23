pub struct DB {}

impl DB {
    pub fn query<'a>(&'a self, q: &'a str) -> &'a str {
        q
    }
}
