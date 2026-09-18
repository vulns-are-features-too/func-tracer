pub struct ReadOnlyDb {}
pub struct ReadWriteDb {}

impl Readonly for ReadOnlyDb {
    fn query<'a>(&'a self, q: &'a str) -> &'a str {
        q
    }
}

impl Readonly for ReadWriteDb {
    fn query<'a>(&'a self, q: &'a str) -> &'a str {
        q
    }
}

impl Writeable for ReadWriteDb {
    fn update<'a>(&'a mut self, q: &'a str) -> &'a str {
        q
    }
}

pub trait Readonly {
    fn query<'a>(&'a self, q: &'a str) -> &'a str;
}

pub trait Writeable: Readonly {
    fn update<'a>(&'a mut self, q: &'a str) -> &'a str;
}
