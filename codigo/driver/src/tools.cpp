#include <cassert>
#include <iostream>
#include <memory>
#include <optional>
#include <unistd.h>

#include "config.hpp"
#include "httpc.hpp"
#include "isclient.hpp"

#include <Magick++.h>

using namespace mifs;

// Config stuff
constexpr std::string_view CMD_SIGNUP_VAL = "signup";
constexpr std::string_view CMD_LOGIN_VAL = "login";
constexpr std::string_view CMD_S2FA_VAL = "2fa";
constexpr std::string_view CMD_LIST_SERVERS_VAL = "list-servers";
constexpr std::string_view CMD_LINK_SERVER_VAL = "link-server";
enum class Command { SIGNUP, LOGIN, S2FA, LIST_SERVERS, LINK_SERVERS };

struct Arguments {
    Command command;
    std::string config_file;
    std::string user_name;
    std::string email;
    std::string password;
    std::string otp;
    std::string organization;
    std::string server;
};

// Forwards decls
std::optional<Arguments> parse_args(int argc, char **argv);
void show_usage();
int signup(const Arguments& args);
int auth(const Arguments& args);
int s2fa(const Arguments& args);
int list_servers(const Arguments& args);
int link_server(const Arguments& args);

int main(int argc, char **argv)
{
    auto args{parse_args(argc, argv)};
    if (!args) {
        return 1;
    }

    switch (args->command) {
    case Command::LOGIN: return auth(*args);
    case Command::S2FA: return s2fa(*args);
    case Command::SIGNUP: return signup(*args);
    case Command::LIST_SERVERS: return list_servers(*args);
    case Command::LINK_SERVERS: return link_server(*args);
    }

    std::cerr << "invalid command code: " << static_cast<int>(args->command) << '\n';
    return 0;
}

int signup(const Arguments& args)
{
    auto config{mifs::Config::parse(args.config_file)};
    assert(config);

    auto client{std::make_shared<http::Client>()};
    auto is_client{mifs::apiclients::IndexServerClient{
        client, mifs::apiclients::IndexServerClient::Config::from_parsed_conf(*config)}};

    auto error{is_client.signin(args.user_name, args.email, args.password)};
    if (!error) {
        return 0;
    }
    std::cerr << "error authenticating user: " << error->message() << '\n';
    return 2;
}

int auth(const Arguments& args)
{
    auto config{mifs::Config::parse(args.config_file)};
    assert(config);

    auto client{std::make_shared<http::Client>()};
    auto is_client{mifs::apiclients::IndexServerClient{
        client, mifs::apiclients::IndexServerClient::Config::from_parsed_conf(*config)}};

    auto resp{is_client.auth(args.email, args.password, args.otp)};
    if (!resp) {
        std::cerr << "error authenticating user: " << resp.error().message() << '\n';
        return 2;
    }

    std::cout << "export MIFS_IS_TOKEN=" << resp->token() << '\n';

    return 0;
}

int s2fa(const Arguments& args)
{
    auto config{mifs::Config::parse(args.config_file)};
    assert(config);

    auto client{std::make_shared<http::Client>()};
    auto is_client{mifs::apiclients::IndexServerClient{
        client, mifs::apiclients::IndexServerClient::Config::from_parsed_conf(*config)}};

    auto resp{is_client.setup2fa()};
    if (!resp) {
        std::cerr << "error setting up 2-factor auth: " << resp.error().message() << '\n';
        return 2;
    }

    Magick::Blob contents{resp->data(), resp->size()};
    Magick::Image myMagickImage{contents};
    myMagickImage.display();
    return 0;
}

int list_servers(const Arguments& args)
{
    auto config{mifs::Config::parse(args.config_file)};
    assert(config);

    auto client{std::make_shared<http::Client>()};
    auto is_client{mifs::apiclients::IndexServerClient{
        client, mifs::apiclients::IndexServerClient::Config::from_parsed_conf(*config)}};

    auto resp{is_client.get_servers()};
    if (!resp) {
        std::cerr << "error getting server list: " << resp.error().message() << '\n';
        return 2;
    }

    const auto& servers{resp->data["servers"]};
    for (const auto& server : servers) {
        std::cout << fmt::format("Id\{}\tOrganization={}\tName={}\n", server.id(), server.org_name(),
                                 server.name());
    }

    return 0;
}

int link_server(const Arguments& args)
{
    auto config{mifs::Config::parse(args.config_file)};
    assert(config);

    auto client{std::make_shared<http::Client>()};
    auto is_client{mifs::apiclients::IndexServerClient{
        client, mifs::apiclients::IndexServerClient::Config::from_parsed_conf(*config)}};

    auto org_it{config->creds().find(args.organization)};
    if (org_it == config->creds().end()) {
        std::cerr << fmt::format("organization '{}' not found.\n", args.organization);
        return 1;
    }

    const auto& servers4org{org_it->second};
    auto server_it{servers4org.find(args.server)};
    if (server_it == servers4org.end()) {
        std::cerr << fmt::format("'{}' not found in organization '{}' server list\n", args.server,
                                 args.organization);
        return 1;
    }

    const auto& creds{server_it->second};
    auto error{is_client.link_fs(args.organization, args.server, creds.client_certificate_fn,
                                 creds.client_private_key_fn, false)};
    if (!error) {
        return 0;
    }

    std::cerr << "error linking file server: " << error->message() << '\n';
    return 2;
}

std::optional<Arguments> parse_args(int argc, char **argv)
{

    if (argc <= 1) {
        show_usage();
        return std::nullopt;
    }

    Arguments args;
    if (argv[1] == CMD_SIGNUP_VAL) {
        args.command = Command::SIGNUP;
    } else if (argv[1] == CMD_LOGIN_VAL) {
        args.command = Command::LOGIN;
    } else if (argv[1] == CMD_S2FA_VAL) {
        args.command = Command::S2FA;
    } else if (argv[1] == CMD_LIST_SERVERS_VAL) {
        args.command = Command::LIST_SERVERS;
    } else if (argv[1] == CMD_LINK_SERVER_VAL) {
        args.command = Command::LINK_SERVERS;
    } else {
        std::cerr << "unknown command: " << argv[1] << '\n';
        return std::nullopt;
    }

    auto _getopt{[&]() { return getopt(argc - 1, &argv[1], "c:u:e:p:o:g:s:"); }};
    for (auto opt{_getopt()}; opt != -1; opt = _getopt()) {
        switch (opt) {
        case 'c': args.config_file = optarg; break;
        case 'u': args.user_name = optarg; break;
        case 'e': args.email = optarg; break;
        case 'p': args.password = optarg; break;
        case 'o': args.otp = optarg; break;
        case 'g': args.organization = optarg; break;
        case 's': args.server = optarg; break;
        case '?': show_usage(); return std::nullopt;
        }
    }

    if (args.config_file.empty()) {
        std::cerr << "config file is mandatory\n";
        return std::nullopt;
    }

    return args;
}

void show_usage()
{
    std::cout
        << "usage: mifs-tools <COMMAND> [options]\n"
        << "\n\tsignup\t(create a new account)\n"
        << "\t\t-c <CONFIG>\t(mandatory)\tPath to the JSON config file with index-server & file server configurations.\n"
        << "\t\t-u <USER>\t(mandatory)\tUsername used in the signup process.\n"
        << "\t\t-e <EMAIL>\t(mandatory)\tEmail to be used to signup & authenticate\n"
        << "\t\t-p <PASSWORD>\t(mandatory)\tPassword to be used to signup & authenticate\n"
        << "\n\tlogin\t(create a session & export credentials -- use with `eval $(mifs-tools login -c ...)\n"
        << "\t\t-c <CONFIG>\t(mandatory)\tPath to the JSON config file with index-server & file server configurations.\n"
        << "\t\t-e <EMAIL>\t(mandatory)\tEmail to be used to signup & authenticate\n"
        << "\t\t-p <PASSWORD>\t(mandatory)\tPassword to be used to signup & authenticate\n"
        << "\t\t-o <OTP>\t(optional)\tMulti-Factor-Auth one-time-password generated by configured device.\n"
        << "\n\t2fa\tSetup multi-factor-auth to allow accessing restricted endpoints.\n"
        << "\t\t-c <CONFIG>\t(mandatory)\tPath to the JSON config file with index-server & file server configurations.\n"
        << "\n\tlist-servers\tList servers the user can link to.\n"
        << "\t\t-c <CONFIG>\t(mandatory)\tPath to the JSON config file with index-server & file server configurations.\n"
        << "\n\tlink-server\tLink an account in a file server.\n"
        << "\t\t-c <CONFIG>\t(mandatory)\tPath to the JSON config file with index-server & file server configurations.\n"
        << "\t\t-g <ORGANIZATION>\t(mandatory)\tOrganization containing the server to link to\n"
        << "\t\t-s <SERVER>\t(mandatory)\tServer to link to\n";
}
