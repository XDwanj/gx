interface Repository {
  fun load(): User
}

class User(val name: String)

object Defaults {
  const val Limit = 10
}

enum class Status {
  Open,
  Closed
}

typealias UserId = String

fun buildUser(): User {
  println("build")
  return User("Ada")
}
