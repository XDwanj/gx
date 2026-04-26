fun buildUser(): User {
  println("build")
  return User("Ada")
}

class User(val name: String)
