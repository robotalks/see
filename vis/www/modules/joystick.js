(function(exports) {
    'use strict';

    function mouseHandler(fn, ctx, elem) {
        return function (evt) {
            return fn.call(ctx,
                vis.relativePos(elem || evt.currentTarget || evt.target, evt.clientX, evt.clientY),
                evt);
        };
    }

    function touchHandler(fn, ctx, elem) {
        return function (evt) {
            if (evt.touches.length == 0) {
                return;
            }
            return fn.call(ctx,
                vis.relativePos(elem || evt.currentTarget || evt.target,
                    evt.touches[0].clientX, evt.touches[0].clientY),
                    evt);
        };
    }

    vis.defineObject('joystick', {
        createContent: function () {
            var cntr = document.createElement('div');
            cntr.classList.add('container');
            cntr.addEventListener('touchstart', touchHandler(this._touchStart, this), true);
            this._touchMoveHandler = touchHandler(this.moving, this, cntr);
            this._touchEndHandler = this._touchEnd.bind(this);
            cntr.addEventListener('mousedown', mouseHandler(this._mouseDown, this), true);
            this._mouseMoveHandler = mouseHandler(this.moving, this, cntr);
            this._mouseUpHandler = mouseHandler(this._mouseUp, this, cntr);
            this._cntr = cntr;

            var stick = document.createElement('div');
            stick.classList.add('stick');
            cntr.appendChild(stick);
            this._stick = stick;

            this._offset = { x: 0, y: 0};
            return cntr;
        },

        place: function () {
            this._stick.style.left = (((this._cntr.clientWidth - this._stick.clientWidth) >> 1) + this._offset.x) + 'px';
            this._stick.style.top = (((this._cntr.clientHeight - this._stick.clientHeight) >> 1) + this._offset.y) + 'px';
        },

        destroy: function () {
            document.removeEventListener('touchmove', this._touchMoveHandler, true);
            document.removeEventListener('touchend', this._touchEndHandler, true);
            document.removeEventListener('mousemove', this._mouseMoveHandler, true);
            document.removeEventListener('mouseup', this._mouseUpHandler, true);
        },

        _touchStart: function (pos, evt) {
            console.log('touchstart', pos, evt)
            document.addEventListener('touchmove', this._touchMoveHandler, true);
            document.addEventListener('touchend', this._touchEndHandler, true);
            this.moveStart(pos);
        },

        _touchEnd: function (evt) {
            this.moveEnd();
            document.removeEventListener('touchmove', this._touchMoveHandler, true);
            document.removeEventListener('touchend', this._touchEndHandler, true);
        },

        _mouseDown: function (pos, evt) {
            document.addEventListener('mousemove', this._mouseMoveHandler, true);
            document.addEventListener('mouseup', this._mouseUpHandler, true);
            this.moveStart(pos);
        },

        _mouseUp: function (pos, evt) {
            this.moveEnd();
            document.removeEventListener('mousemove', this._mouseMoveHandler, true);
            document.removeEventListener('mouseup', this._mouseUpHandler, true);
        },

        moveStart: function (pos) {
            this._startPos = pos;
        },

        moving: function (pos) {
            var offMax = this.offsetMax();
            var off = {
                x: pos.x - this._startPos.x,
                y: pos.y - this._startPos.y
            };
            if (Math.abs(off.x) > offMax.w) {
                off.x = off.x > 0 ? offMax.w : -offMax.w;
            }
            if (Math.abs(off.y) > offMax.h) {
                off.y = off.y > 0 ? offMax.h : -offMax.h;
            }
            this.updateOffset(off);
        },

        moveEnd: function () {
            this.updateOffset({ x: 0, y: 0 });
        },

        offsetMax: function () {
            return {
                w: (this._cntr.clientWidth - this._stick.clientWidth) >> 1,
                h: (this._cntr.clientHeight - this._stick.clientHeight) >> 1
            };
        },

        updateOffset: function (off) {
            if (this.properties.x === false) {
                off.x = 0;
            }
            if (this.properties.y === false) {
                off.y = 0;
            }
            this._offset = off;
            this.place();

            var rc = this.measure();
            var offMax = this.offsetMax();

            var pos = {
                x: off.x / offMax.w,
                y: - off.y / offMax.h,
            };
            var rel = {
                x: rc.w * pos.x / 2,
                y: rc.h * pos.y / 2
            };
            this.emit({
                action: 'stick',
                stick: {
                    id: this.id(),
                    data: this.properties.data,
                    pos: pos,
                    rel: rel,
                }
            });
        }
    });
})(window);
