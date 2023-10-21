<!DOCTYPE html>
<html lang="en" class="no-min-dimensions">
<head>
    <meta http-equiv="Content-Type" content="text/html; charset=UTF-8"/>
    <title>Forgot your password?</title>

    <link rel="icon" type="image/x-icon" href="{{.StaticPath}}/favicon.png">
    <link rel="apple-touch-icon" href="{{.StaticPath}}/apple-touch-icon.png?">
    <link rel="apple-touch-icon-precomposed" href="{{.StaticPath}}/apple-touch-icon.png">
    <link rel="mask-icon" href="{{.StaticPath}}/ninja-portrait.svg" color="#6078FF">

    <meta charset="utf-8" />
    <meta http-equiv="X-UA-Compatible" content="IE=edge,chrome=1">
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <meta name="apple-mobile-web-app-capable" content="yes">
    <meta name="apple-mobile-web-app-status-bar-style" content="black">
    <meta name="apple-mobile-web-app-title" content="Password reset">

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
          <img alt="" src="{{.StaticPath}}/ninja-portrait.svg" />
        </div>

        <h1 class="onboarding__title">
          Reset your password
        </h1>
      </header>

      <form class="simple_form onboarding__form" id="new_user" novalidate="novalidate" action="/auth/reset" accept-charset="UTF-8" method="post">
        <div class="form-group hidden user_reset_password_token"><input class="form-control hidden" type="hidden" name="user[reset_password_token]" id="user_reset_password_token" /></div>

        <div class="onboarding__field onboarding__field--hide-label">
          <label class="onboarding__label" id="reset-password">
            Email Address
          </label>
          <input class="string email required onboarding__input required js-onboarding-email" id="reset-password" required="required" autocomplete="username" aria-required="true" placeholder="Email Address" type="email" value="" name="user[email]" />
        </div>

        <div class="onboarding__actions">
          <button class="button onboarding__button onboarding__button--full-width">
            Reset password
          </button>
        </div>
        </form>
    </div>
  </main>
</section>

</body>
</html>
