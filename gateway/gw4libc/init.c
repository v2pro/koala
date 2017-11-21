static int _init;
static int _is_tracing;

void go_initialized(int is_tracing) {
    _init = 1;
    _is_tracing = is_tracing;
}

int is_go_initialized() {
    return _init;
}

int is_tracing() {
    return _is_tracing;
}