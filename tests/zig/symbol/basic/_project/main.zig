const User = struct {
    id: i32,
};

pub fn buildUser() User {
    return .{ .id = 1 };
}
