<!DOCTYPE html>
<html lang="en" class="no-min-dimensions">
<head>
    <meta http-equiv="Content-Type" content="text/html; charset=UTF-8"/>

    <title>Login</title>

    <link rel="icon" type="image/x-icon" href="{{.StaticPath}}/favicon.png">

    <link rel="apple-touch-icon" href="{{.StaticPath}}/apple-touch-icon.png">
    <link rel="apple-touch-icon-precomposed" href="{{.StaticPath}}/apple-touch-icon.png">
    <link rel="mask-icon" href="{{.StaticPath}}/ninja-portrait.svg" color="#6078FF">

    <meta charset="utf-8" />
    <meta http-equiv="X-UA-Compatible" content="IE=edge,chrome=1">
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <meta name="apple-mobile-web-app-capable" content="yes">
    <meta name="apple-mobile-web-app-status-bar-style" content="black">
    <meta name="apple-mobile-web-app-title" content="Login">

    <link rel="preload" href="{{.StaticPath}}/fa-regular-400.woff2" as="font" type="font/woff2" crossorigin="anonymous">
    <link rel="preload" href="{{.StaticPath}}/fa-solid-900.woff2" as="font" type="font/woff2" crossorigin="anonymous">
    <link rel="preload" href="{{.StaticPath}}/Inter-Regular.woff2" as="font" type="font/woff2" crossorigin="anonymous">
    <link rel="preload" href="{{.StaticPath}}/Inter-SemiBold.woff2" as="font" type="font/woff2" crossorigin="anonymous">

    <link rel="stylesheet" media="screen" href="{{.StaticPath}}/application.css" />
</head>
<body class="no-min-dimensions">

<section class="onboarding onboarding--centered">
    <main class="onboarding__main">
        <div class="onboarding__wrapper">

            <header class="onboarding__header">
                <div class="onboarding__logo">
                    <img alt="" src="{{.StaticPath}}/ninja-portrait.svg" color="#6078FF"/>
                </div>

                <h1 class="onboarding__title">Sign in</h1>
            </header>

            <form class="button_to" method="get" action="{{.ProviderLoginPath}}">
                <input class="button onboarding__button onboarding__button--full-width onboarding__button--google" type="submit" value="Continue with Google" />
            </form>

            <p class="onboarding__options-separator">
                or sign in using email
            </p>

            <form class="simple_form onboarding__form" id="new_user" novalidate="novalidate" action="{{.LoginPath}}" accept-charset="UTF-8" method="post">
                <div class="form-group hidden user_remember_me"><input class="form-control hidden" type="hidden" value="1" name="user[remember_me]" id="user_remember_me" /></div>


                <div class="onboarding__field onboarding__field--hide-label">
                    <label class="onboarding__label" for="signin-email-address">
                        Email address
                    </label>

                    <input class="string email required onboarding__input js-onboarding-email" id="signin-email-address" required="required" autofocus="autofocus" autocomplete="username" aria-required="true" placeholder="Email address" type="email" name="user[email]" />
                </div>

                <div class="onboarding__field onboarding__field--hide-label">
                    <label class="onboarding__label" id="signin-password">
                        Password
                    </label>

                    <input class="password required onboarding__input required" id="signin-password" required="required" autocomplete="current-password" aria-required="true" placeholder="Password" type="password" name="user[password]" />
                </div>

                <div class="onboarding__actions">
                    <button class="button onboarding__button onboarding__button--full-width">
                        Sign in
                    </button>
                </div>
            </form>
            <p class="onboarding__footer">
                <a href="/users/password/new">Forgot your password?</a>
            </p>
        </div>
    </main>
</section>

</body>
</html>
