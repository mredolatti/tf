#ifndef MIFS_LOG_HPP
#define MIFS_LOG_HPP

#include "buildcfg.hpp"
#include <spdlog/logger.h>
#include <spdlog/spdlog.h>

namespace mifs::log
{

using logger_t = std::shared_ptr<spdlog::logger>;

logger_t initialize();
logger_t get();

}; // namespace mifs::log

#endif
