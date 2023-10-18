use std::{error::Error, net::ToSocketAddrs};

fn main() -> Result<(), Box<dyn Error>> {
    println!("spawning thread to lookup localhost ...");
    let t = std::thread::spawn(lookup_localhost);

    const NUM_ITERATIONS: usize = 100;
    for i in 0..NUM_ITERATIONS {
        std::env::set_var(format!("ENV_VAR_{i}"), "value");
    }

    t.join().expect("BUG thread must succeed");
    Ok(())
}

fn lookup_localhost() {
    let addrs_iter = "localhost:1".to_socket_addrs().unwrap();
    for addr in addrs_iter {
        println!("localhost: ip={} port={}", addr.ip(), addr.port());
    }
}
