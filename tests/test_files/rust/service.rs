use crate::db::DB;

pub struct Service {
    i: i8,
    db: DB,
}

impl Service {
    pub fn new() -> Service {
        Service { i: 0, db: DB {} }
    }

    pub fn foo(&self) {
        self.db.query("foo");
    }

    pub fn bar(&mut self) {
        self.i += 1;
        self.db.query("bar");
    }
}
