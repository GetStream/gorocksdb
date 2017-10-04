extern crate libc;
extern crate core;
use std::ffi::CStr;
use std::ffi;
use core::slice;
mod rocksdb;
use rocksdb::*;
use std::mem;
use std::iter;

#[repr(C)]
pub struct rust_rocksdb_many_keys_t {
    keys: *mut *mut libc::c_char,
    key_sizes: *mut libc::size_t,
    values: *mut *mut libc::c_char,
    value_sizes: *mut libc::size_t,
    found: libc::uintptr_t,
}

#[repr(C)]
pub struct rust_rocksdb_many_keys_filter_t {
    key_prefix : *const libc::c_char,
    key_prefix_size: libc::size_t,
    key_stop : *const libc::c_char,
    key_stop_size: libc::size_t,
}

pub unsafe extern "C" fn rust_rocksdb_destroy_many_keys(many_keys : rust_rocksdb_many_keys_t) {
    let len = many_keys.found;
    let keys = Box::from_raw(slice::from_raw_parts_mut(many_keys.keys, len));
    let key_sizes = Box::from_raw(slice::from_raw_parts_mut(many_keys.key_sizes, len));
    let values = Box::from_raw(slice::from_raw_parts_mut(many_keys.values, len));
    let value_sizes = Box::from_raw(slice::from_raw_parts_mut(many_keys.value_sizes, len));
    for i in 0..len {
        Box::from_raw(slice::from_raw_parts_mut(keys[i], key_sizes[i]));
        Box::from_raw(slice::from_raw_parts_mut(values[i], value_sizes[i]));
    }
}



unsafe fn boxed_slice_from_raw_parts<T: Copy>(p: *const T, len: usize) -> Box<[T]> {
    // Create a slice
    let slice = unsafe { slice::from_raw_parts(p, len) };
    // Copy the slice on the heap
    slice.into()
}

unsafe fn option_slice_from_raw_parts<'a, T>(p: *const T, len: usize) -> Option<&'a[T]> {
    if len > 0 {
        Some(unsafe { slice::from_raw_parts(p, len) })
    } else {
        None
    }
}


#[no_mangle]
pub extern "C" fn rust_rocksdb_iter_next_many_keys_f(
    iter: *mut rocksdb_iterator_t,
    limit: libc::int32_t,
    key_filter: rust_rocksdb_many_keys_filter_t,
) -> rust_rocksdb_many_keys_t {
    let mut keys: Vec<*mut libc::c_char> = vec![];
    let mut key_sizes: Vec<libc::size_t> = vec![];

    let mut values: Vec<*mut libc::c_char> = vec![];
    let mut value_sizes: Vec<libc::size_t> = vec![];

    let prefix = unsafe {
        option_slice_from_raw_parts(key_filter.key_prefix, key_filter.key_prefix_size)
    };

    let stop = unsafe {
        option_slice_from_raw_parts(key_filter.key_stop, key_filter.key_stop_size)
    };


    while unsafe{rocksdb_iter_valid(iter)} != 0 {
        let mut key_size: libc::size_t = 0;
        let raw_key : *const libc::c_char = unsafe {rocksdb_iter_key(iter, &mut key_size)};
        let mut key = unsafe {boxed_slice_from_raw_parts(raw_key, key_size) };

        if let Some(prefix) = prefix {
            if !key.starts_with(prefix) {
                break
            }
        }
        if let Some(stop) = stop {
            if !key.starts_with(stop) {
                break
            }
        }


        key_sizes.push(key_size);
        keys.push(key.as_mut_ptr());


        let mut value_size: libc::size_t = 0;
        let raw_value : *const libc::c_char = unsafe {rocksdb_iter_value(iter, &mut value_size)};
        let mut value = unsafe {boxed_slice_from_raw_parts(raw_value, value_size) };
        value_sizes.push(value_size);
        values.push(value.as_mut_ptr());

        mem::forget(key);
        mem::forget(value);

        unsafe{rocksdb_iter_next(iter)};

        if limit > 0 && keys.len() as i32 == limit {
            break
        }

    }



    return rust_rocksdb_many_keys_t{
        found: keys.len(),
        keys: vec_to_ptr(keys),
        key_sizes: vec_to_ptr(key_sizes),
        values: vec_to_ptr(values),
        value_sizes: vec_to_ptr(value_sizes),
    }
}

fn vec_to_ptr<T>(v: Vec<T>) -> *mut T {
    let mut boxed = v.into_boxed_slice();
    let pointer = boxed.as_mut_ptr();
    mem::forget(boxed);
    return pointer
}
