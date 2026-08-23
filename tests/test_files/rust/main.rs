mod chain;
mod db;
mod external;
mod nested;
mod recurse;
mod service;

use service::Service;

fn main() {
    start();

    let mut svc = Service::new();
    svc.foo();
    svc.bar();

    recurse::recurse_self(2);
    recurse::recurse_other(5);

    external::ext_hello();
    external::ext_print("hi");

    nested::nested0();

    _ = chain::chain(true) || chain::chain(false);

    end();
}

fn start() {}

fn end() {}
