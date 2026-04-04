const LIMIT: i32 = 10;

struct User;

enum Status {
    Open,
    Closed,
}

trait Reader {
    fn read(&self);
}

type UserId = i64;

mod billing {}

fn build_user() {}

impl User {
    fn load(&self) {}
}
