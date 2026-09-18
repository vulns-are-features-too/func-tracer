use crate::db::*;

pub struct Service {
    ro_db: ReadOnlyDb,
    rw_db: ReadWriteDb,
}

pub fn run() {
    let mut svc = Service::new();
    svc.foo();
    svc.bar();
    svc.save();
}

impl Service {
    pub fn new() -> Service {
        Service {
            ro_db: ReadOnlyDb {},
            rw_db: ReadWriteDb {},
        }
    }

    pub fn foo(&self) {
        self.ro_db.query("foo");
    }

    pub fn bar(&mut self) {
        self.ro_db.query("bar");
        self.rw_db.query("bar");
    }

    pub fn save(&mut self) {
        self.rw_db.update("save");
    }
}
