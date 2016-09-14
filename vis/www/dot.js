(function (exports) {
    'use strict';

    vis.defineCanvasObject('dot', {
        paint: function (rc, canvas) {
            var ctx = canvas.getContext('2d');
            ctx.clearRect(0, 0, rc.w, rc.h);
            var x0 = rc.w >> 1;
            var y0 = rc.h >> 1;
            var r = Math.floor(Math.min(rc.w, rc.h) / 2);
            ctx.fillStyle = $(canvas).css('color');
            ctx.beginPath();
            ctx.arc(x0, y0, r, 0, 2 * Math.PI);
            ctx.fill();
        }
    });
})(window);
