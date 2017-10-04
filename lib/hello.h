void hello(char *name);

typedef struct rust_rocksdb_many_keys_filter_t{
    const char* key_prefix;
    size_t key_prefix_size;
    const char* key_stop;
    size_t key_stop_size;

} rust_rocksdb_many_keys_filter_t;

typedef struct rust_rocksdb_many_keys_t {
    char** keys;
    size_t* key_sizes;
    char** values;
    size_t* value_sizes;
    int found;

} rust_rocksdb_many_keys_t;

rust_rocksdb_many_keys_t rust_rocksdb_iter_next_many_keys_f(rocksdb_iterator_t* iter, int limit, rust_rocksdb_many_keys_filter_t filter);
