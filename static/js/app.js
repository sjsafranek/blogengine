
var UI = function() {
    this._enableToTop();
    this._enableNavBar();
}

UI.prototype._enableToTop = function () {

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
        );

    $('body').append($toTop);

    $(window).scroll(function () {
        if ($(this).scrollTop() > 50) {
            $toTop.fadeIn();
        } else {
            $toTop.fadeOut();
        }
    });

    $toTop.click(function () {
        $('body').animate({scrollTop:0}, 'slow');
    });
},

UI.prototype._enableNavBar = function() {
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


var App = function() {
    this.ui = new UI();
}


var app;

$(document).ready(function(){

    app = new App();

    // Build sidebar achors
    // Enable scrollspy
    $('.media-body h1, .media-body h2, .media-body h3, .media-body h4, .media-body h5, .media-body h6').each(function(i, elem) {
        var $elem = $(elem);
        $elem.append(
            $('<a>',{id: 'ss'+i}).addClass('anchor')
        );
        $('#menu').append(
            $('<li>').addClass('nav-item')
                .append(
                    $('<a>', {href: '#ss'+i}).addClass('nav-link').append($elem.text())
                )
        );
    });
    $('body').scrollspy({ target: '#menu' });
    $('[data-spy="scroll"]').each(function () {
        $(this).scrollspy('refresh');
    });



    // Post Carousel
    $($('.carousel-item')[0]).addClass('active');

    var $elem = $('#postCarousel').carousel({
        'interval': 5000
    });

    $('.post-link').on('click', function(e) {
        window.location.href = $(e.target).closest('.post-link').attr('href');
    });


});
