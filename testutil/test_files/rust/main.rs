mod chain;
mod db;
mod external;
mod nested;
mod recurse;
mod service;

fn main() {
    start();

    service::run();

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
