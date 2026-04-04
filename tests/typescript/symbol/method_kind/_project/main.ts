namespace billing {}

class User {
  method(): number {
    return 1;
  }
}

interface Reader {
}

enum Status {
  Open,
  Closed,
}

type UserId = number;

const LIMIT = 10;

function buildUser(): User {
  return new User();
}
