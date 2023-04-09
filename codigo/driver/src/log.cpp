#include "log.hpp"
#include "spdlog/sinks/stdout_color_sinks.h"
#include <spdlog/common.h>
#include <spdlog/logger.h>
#include <spdlog/spdlog.h>

namespace mifs::log
{

static const std::string logger_name{"mifs"};

logger_t initialize()
{
    auto logger{spdlog::stdout_color_mt(logger_name)};
    logger->set_level(spdlog::level::trace);
    return logger;
}

logger_t get() { return spdlog::get(logger_name); }

} // namespace mifs::log
