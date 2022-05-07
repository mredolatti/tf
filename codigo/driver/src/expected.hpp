#ifndef MIFS_UTIL_EXPECTED
#define MIFS_UTIL_EXPECTED

#include <initializer_list>
#include <type_traits>
#include <utility>
#include <functional>

namespace mifs::util {

template<typename ErrorType>
class Unexpected
{
    private:
    ErrorType error_;

    public:
    explicit Unexpected(ErrorType rhs) : error_{std::move(rhs)} {}

    Unexpected() = delete;
    Unexpected(const Unexpected<ErrorType>&) = delete;
    Unexpected& operator=(const Unexpected<ErrorType>&) = delete;
    Unexpected(Unexpected<ErrorType>&&) noexcept = default;
    Unexpected& operator=(Unexpected<ErrorType>&&) noexcept = default;
    ~Unexpected() = default;

    ErrorType& error() & { return error_; }
    const ErrorType& error() const& { return error_; }
    ErrorType&& error() && { return error_; }
};

template<typename ResultType, typename ErrorType>
class Expected
{
    public:
    using result_t = ResultType;
    using error_t = ErrorType;

    private:
    union
    {
        result_t result_;
        error_t error_;
    };
    bool ok_{false};

    public:
    Expected() = delete;

    template<typename U = result_t, typename = typename std::enable_if<
        !std::is_same<Expected<result_t, error_t>, typename std::decay<U>::type>::value
        && !std::is_base_of<Expected<result_t, error_t>, U>::value>::type>
    explicit Expected(U&& rhs) : ok_{true}
    {
        new (&result_) result_t(std::forward<U>(rhs));
    }

    Expected(const result_t&);              // NOLINT(google-explicit-constructor)
    Expected(const Unexpected<error_t>&);   // NOLINT(google-explicit-constructor)
    Expected(Unexpected<error_t>&&);        // NOLINT(google-explicit-constructor)
    Expected(const Expected& rhs);          // NOLINT(google-explicit-constructor)
    Expected(Expected&& rhs) noexcept;      // NOLINT(google-explicit-constructor)
    Expected& operator=(const Expected&);
    Expected& operator=(Expected&&) noexcept;
    ~Expected();

    explicit operator bool() const { return ok_; }

    result_t&         operator*() &      { return result_; }
    result_t&&        operator*() &&     { return std::move(result_); }
    const result_t&   operator*() const& { return result_; }

    error_t& error() & { return error_; }
    const error_t& error() const& { return error_; }
    error_t&& error() && { return error_; }

    template<typename Fallback>
    result_t result_or(Fallback&& fallback) const&;

    template<typename Fallback>
    result_t result_or(Fallback&& fallback) &&;

    template<typename S = result_t, typename E = error_t>
    typename std::enable_if<
        std::is_nothrow_move_constructible<S>::value
        && std::is_nothrow_constructible<E>::value
    >::type
    swap(Expected<result_t, error_t>& rhs)
    {
        if (ok_ && rhs.ok_) {
            using std::swap; 
            swap(result_, rhs.result_); 
        } else if (!ok_ && !rhs.ok_) {
            using std::swap;
            swap(error_, rhs.error_);
        } 
        else if (ok_ && !rhs.ok_) {
            rhs.swap(*this); // recursive call to this function to invert the swapee. Case implemented below
        } else { // !ok && rhs.ok_
            ErrorType my_error{std::move(error_)};
            error_.~ErrorType();
            new (&result_) ResultType(std::move(rhs.result_));
            ok_ = true;
            rhs.result_.~ResultType();
            new (&rhs.error_) ErrorType(std::move(my_error));
            rhs.ok_ = false;
        }
    }
};

template<class ResultType, typename ErrorType>
Expected<ResultType, ErrorType>::Expected(const ResultType& rhs)
    : ok_{true}
{
    new (&result_) ResultType(rhs);
}

template<class ResultType, typename ErrorType>
Expected<ResultType, ErrorType>::Expected(const Unexpected<ErrorType>& err) 
{
    new (&error_) ErrorType(err);
}

template<class ResultType, typename ErrorType>
Expected<ResultType, ErrorType>::Expected(Unexpected<ErrorType>&& err) {
    new (&error_) ErrorType(std::move(err.error()));
}

template<typename ResultType, typename ErrorType>
Expected<ResultType, ErrorType>::~Expected()
{
    if (ok_) {
        result_.~ResultType();
    } else {
        error_.~ErrorType();
    }
}

template<typename ResultType, typename ErrorType>
Expected<ResultType, ErrorType>::Expected(const Expected& rhs) :
    ok_{rhs.ok_}
{
    if (rhs.ok_) {
        new (&result_) ResultType(rhs.result_);
    } else {
        new (&error_) ErrorType(rhs.error_);
    }
}

template<typename ResultType, typename ErrorType>
Expected<ResultType, ErrorType>::Expected(Expected&& rhs) noexcept :
    ok_{rhs.ok_}
{
    if (rhs.ok_) {
        new (&result_) ResultType(std::move(rhs.result_));
    } else {
        new (&error_) ErrorType(std::move(rhs.error_));
    }
}

template<typename ResultType, typename ErrorType>
Expected<ResultType, ErrorType>& Expected<ResultType, ErrorType>::operator=(const Expected& rhs)
{
    Expected<ResultType, ErrorType> tmp{rhs};
    tmp.swap(*this);
    return *this;
}

template<typename ResultType, typename ErrorType>
Expected<ResultType, ErrorType>& Expected<ResultType, ErrorType>::operator=(Expected&& rhs) noexcept
{
    Expected<ResultType, ErrorType> tmp{std::move(rhs)};
    tmp.swap(*this);
    return *this;
}

template<typename ResultType, typename ErrorType>
template<typename Fallback>
typename Expected<ResultType, ErrorType>::result_t Expected<ResultType, ErrorType>::result_or(Fallback&& fallback) const&
{
    static_assert(std::is_copy_constructible<result_t>::value, "expected value should be copy constructible");
    static_assert(std::is_convertible<Fallback&&, result_t>::value, "expected value should be copy constructible");
    return ok_ ? result_ : static_cast<result_t>(std::forward<Fallback>(fallback));
}

template<typename ResultType, typename ErrorType>
template<typename Fallback>
typename Expected<ResultType, ErrorType>::result_t Expected<ResultType, ErrorType>::result_or(Fallback&& fallback) &&
{
    static_assert(std::is_move_constructible<result_t>::value, "expected value should be move constructible");
    static_assert(std::is_convertible<Fallback&&, result_t>::value, "expected value should be copy constructible");
    return ok_ ? std::move(result_) : static_cast<result_t>(std::forward<Fallback>(fallback));
}

} // namespace util

#endif
