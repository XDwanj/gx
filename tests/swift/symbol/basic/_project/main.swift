struct User {
    let id: Int

    init(id: Int) {
        self.id = id
    }
}

func buildUser() -> User {
    return User(id: 1)
}
