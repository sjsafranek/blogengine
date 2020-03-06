
const UI = function() {
    this._enableToTop();
    this._enableNavBar();
    this._enableScrollSpy();
    this._enableCarousel();
}

UI.prototype._enableToTop = function () {
    // create a button element to allow user to
    // scroll to top of page.
    var $toTop = $('<a>', { href: '#'})
        .addClass('btn btn-light')
        .css({
            'position': 'fixed',
            'bottom': '25px',
            'right': '25px',
            'display': 'none'
        })
        .append(
            $('<i>').addClass('fas fa-chevron-up')
        )
        .click(function () {
            $('body').animate({scrollTop:0}, 'slow');
        });

    $('body').append($toTop);

    // add event listener for page scrolling to
    // control the visibility of button element.
    $(window).scroll(function () {
        if ($(this).scrollTop() > 50) {
            $toTop.fadeIn();
        } else {
            $toTop.fadeOut();
        }
    });
},

UI.prototype._enableNavBar = function() {
    //
    var toggleAffix = function(affixElement, scrollElement) {
        var height = affixElement.outerHeight()
        if (scrollElement.scrollTop() > 10){
            affixElement.addClass("affix");
        } else {
            affixElement.removeClass("affix");
        }
    };

    $('[data-toggle="affix"]').each(function() {
        var $elem = $(this);
        $(window).on('scroll resize', function(e) {
            toggleAffix($elem, $(this));
        });
        toggleAffix($elem, $(window));
    });
}

UI.prototype._enableScrollSpy = function() {
    // loop through post body and find all "heading" tags
    $('.media-body h1, .media-body h2, .media-body h3, .media-body h4, .media-body h5, .media-body h6').each(function(i, elem) {
        var $elem = $(elem);
        // add an anchor element to the heading element
        // this helps the scroll behavior keep the heading
        // element in view.
        $elem.append(
            $('<a>',{id: 'ss'+i}).addClass('anchor')
        );
        // add a navigation link to the sidebar menu
        $('#menu').append(
            $('<li>').addClass('nav-item')
                .append(
                    $('<a>', {href: '#ss'+i}).addClass('nav-link').append($elem.text())
                )
        );
    });
    // Initialize the scrollspy behavior
    $('body').scrollspy({ target: '#menu' });
}

UI.prototype._enableCarousel = function() {
    // Initial active element required
    // https://getbootstrap.com/docs/4.0/components/carousel/
    $($('.carousel-item')[0]).addClass('active');
    // Initialize carousel
    var $elem = $('#postCarousel').carousel({
        'interval': 5000
    });
    // Handle page changes via carousel items
    $('.post-link').on('click', function(e) {
        window.location.href = $(e.target).closest('.post-link').attr('href');
    });
}


const App = function() {
    this.ui = new UI();
}


var app;
$(document).ready(function(){
    app = new App();
});
