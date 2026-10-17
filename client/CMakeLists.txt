cmake_minimum_required(VERSION 3.16)
project(snekproxy_client LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

find_package(Qt6 6.2 REQUIRED COMPONENTS Network Widgets)

add_executable(snekproxy-client main.cpp platform_proxy.cpp platform_proxy.h)
target_link_libraries(snekproxy-client PRIVATE Qt6::Network Qt6::Widgets)

if(WIN32)
    target_link_libraries(snekproxy-client PRIVATE wininet)
endif()
