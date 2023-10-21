<!DOCTYPE html>
<html lang="en" class="no-min-dimensions">
<head>
  <meta http-equiv="Content-Type" content="text/html; charset=UTF-8"/>

  <title>Sign up</title>

  <link rel="icon" type="image/x-icon" href="{{.StaticPath}}/favicon.png">

  <link rel="apple-touch-icon" href="{{.StaticPath}}/apple-touch-icon.png">
  <link rel="apple-touch-icon-precomposed" href="{{.StaticPath}}/apple-touch-icon.png">
  <link rel="mask-icon" href="{{.StaticPath}}/ninja-portrait.svg?" color="#6078FF">

  <meta charset="utf-8" />
  <meta http-equiv="X-UA-Compatible" content="IE=edge,chrome=1">
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <meta name="apple-mobile-web-app-capable" content="yes">
  <meta name="apple-mobile-web-app-status-bar-style" content="black">
  <meta name="apple-mobile-web-app-title" content="Signup">

  <link rel="stylesheet" media="screen" href="{{.StaticPath}}/application.css" />
</head>
<body class="no-min-dimensions">

<section class="onboarding" data-provider="tbd">
  <main class="onboarding__main">
    <div class="onboarding__wrapper">
      <header class="onboarding__header">
        <div class="onboarding__logo">
          <img alt="" src="{{.StaticPath}}/ninja-portrait.svg" />
        </div>

        <h1 class="onboarding__title">
          Signup and become a ninja
        </h1>
      </header>

      <form class="button_to" method="post" action="/users/auth/google_oauth2"><input class="button onboarding__button onboarding__button--full-width onboarding__button--google" type="submit" value="Sign up with Google" /></form>

      <p class="onboarding__options-separator">
        or sign up using email
      </p>

      <form class="simple_form onboarding__form" id="new_user" data-rewardful="true" novalidate="novalidate" action="/users" accept-charset="UTF-8" method="post">
          <input type="hidden" class="onboarding__input" name="product_types[]" value="metrics" id="product-metrics">
        <div class="onboarding__field onboarding__field--hide-label js-signup-field">
          <label class="onboarding__label" for="signup-name">
            Full name
          </label>

          <input class="string required onboarding__input" id="signup-name" required="required" aria-required="true" placeholder="Full name" type="text" name="user[name]" />
        </div>

        <div class="onboarding__field onboarding__field--hide-label js-signup-field">
          <label class="onboarding__label" for="signup-email-address">
            Email address
          </label>

          <input class="string email required onboarding__input js-signup-user-email" id="signup-email-address" type="email" required="required" autocomplete="username" aria-required="true" placeholder="Email address" value="" name="user[email]" />

          <div class="onboarding__error js-signup-error">
            That is an odd and invalid email address. Take another look.
          </div>
        </div>

        <div class="onboarding__field onboarding__field--hide-label js-signup-field" id="signup-phone-number-wrapper" style="display: none">
          <label class="onboarding__label" for="signup-phone-number">
            Phone number
          </label>

          <input class="string tel optional onboarding__input js-signup-user-phone" id="signup-phone-number" type="tel" placeholder="Phone number" name="user[phone]" />
        </div>

        <div class="onboarding__field onboarding__field--hide-label js-signup-field">
          <label class="onboarding__label" for="signup-company-name">
            Company name
          </label>

          <input class="string optional onboarding__input js-signup-user-company" id="signup-company-name" placeholder="Company name" type="text" name="user[company]" />
        </div>

        <div class="onboarding__field onboarding__field--hide-label js-signup-field">
          <label class="onboarding__label" id="signup-password">
            Password
          </label>

          <input class="password optional onboarding__input js-signup-user-password required" id="signup-password" autocomplete="new-password" placeholder="Password (at least 8 characters long)" type="password" name="user[password]" />

          <div class="onboarding__error js-signup-error">
            Whoa, that's a short password. Try at least <strong>8</strong> strong characters.
          </div>
        </div>

        <div class="onboarding__field onboarding__field--inline-checkbox">
          <input type="checkbox" class="onboarding__input user-agree" name="tandc" id="tandc" tabindex="0">

          <label class="onboarding__label" for="tandc">
            <span>I agree to the <a href="{{.HomepageURL}}/terms" target="_blank" tabindex="-1">Terms of Service</a> and <a href="{{.HomepageURL}}/privacy" target="_blank" tabindex="-1">Privacy&nbsp;Policy</a>.</span>
          </label>
        </div>

        <div class="onboarding__actions">
          <script src="https://www.recaptcha.net/recaptcha/api.js" async defer ></script>
        <script>
          var invisibleRecaptchaSubmit = function () {
            var closestForm = function (ele) {
              var curEle = ele.parentNode;
              while (curEle.nodeName !== 'FORM' && curEle.nodeName !== 'BODY'){
                curEle = curEle.parentNode;
              }
              return curEle.nodeName === 'FORM' ? curEle : null
            };

            var el = document.querySelector(".g-recaptcha")
            if (!!el) {
              var form = closestForm(el);
              if (form) {
                form.submit();
              }
            }
          };
        </script>

        <button type="submit" data-sitekey="6Le5oJgUAAAAAHEUPnQyE64Jynhg-6ybqh3sMk6f" data-callback="invisibleRecaptchaSubmit" class="g-recaptcha button onboarding__button onboarding__button--full-width" disabled="true">Start your trial</button>

        </div>
        </form>
    </div>
  </main>

</section>

</body>
</html>
