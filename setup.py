from distutils.core import setup, Extension
from os import path

pwd = lambda f: path.join(path.abspath(path.dirname(__file__)), f)
contents = lambda f: open(pwd(f)).read().strip()

module = Extension('inbloom',
    ['inbloom/inbloom.c', 'vendor/libbloom/bloom.c', 'vendor/libbloom/murmur2/MurmurHash2.c'],
    include_dirs=['vendor/libbloom/murmur2']
)

setup(
    name='inbloom',
    author='EverythingMe',
    description=contents('README'),
    version=contents('VERSION'),
    ext_modules=[module]
)
