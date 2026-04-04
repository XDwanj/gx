typedef int Count;

namespace billing {
class Service {
public:
  int Load();
};

struct User {
  int id;
};

enum Status {
  StatusOpen,
  StatusClosed
};

int Service::Load() {
  return 1;
}

int BuildUser() {
  return 1;
}
}
