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
            var touch = null;
            for (var i in evt.touches) {
                if (evt.touches[i].target === elem) {
                    touch = evt.touches[i];
                    break;
                }
            }
            if (touch == null) {
                return;
            }
            return fn.call(ctx,
                vis.relativePos(elem || evt.currentTarget || evt.target,
                    touch.clientX, touch.clientY),
                    evt);
        };
    }

    var listenMode = {capture: true, passive: true};

    function startListen(event, handler) {
        document.addEventListener(event, handler, listenMode);
    }

    function stopListen(event, handler) {
        document.removeEventListener(event, handler, listenMode);
    }

    vis.defineObject('joystick', {
        createContent: function () {
            var cntr = document.createElement('div');
            cntr.classList.add('container');
            this._cntr = cntr;
            var stick = document.createElement('div');
            stick.classList.add('stick');
            cntr.appendChild(stick);
            this._stick = stick;

            cntr.addEventListener('touchstart', touchHandler(this._touchStart, this, stick), true);
            this._touchMoveHandler = touchHandler(this.moving, this, stick);
            this._touchEndHandler = this._touchEnd.bind(this);
            cntr.addEventListener('mousedown', mouseHandler(this._mouseDown, this), listenMode);
            this._mouseMoveHandler = mouseHandler(this.moving, this, cntr);
            this._mouseUpHandler = mouseHandler(this._mouseUp, this, cntr);

            this._offset = { x: 0, y: 0};
            return cntr;
        },

        place: function () {
            this._stick.style.left = (((this._cntr.clientWidth - this._stick.clientWidth) >> 1) + this._offset.x) + 'px';
            this._stick.style.top = (((this._cntr.clientHeight - this._stick.clientHeight) >> 1) + this._offset.y) + 'px';
        },

        destroy: function () {
            stopListen('touchmove', this._touchMoveHandler);
            stopListen('touchend', this._touchEndHandler);
            stopListen('mousemove', this._mouseMoveHandler);
            stopListen('mouseup', this._mouseUpHandler);
        },

        _touchStart: function (pos, evt) {
            if (evt.touches.length > 1) {
                evt.preventDefault();
            }
            startListen('touchmove', this._touchMoveHandler);
            startListen('touchend', this._touchEndHandler);
            this.moveStart(pos);
        },

        _touchEnd: function (evt) {
            this.moveEnd();
            stopListen('touchmove', this._touchMoveHandler);
            stopListen('touchend', this._touchEndHandler);
        },

        _mouseDown: function (pos, evt) {
            startListen('mousemove', this._mouseMoveHandler);
            startListen('mouseup', this._mouseUpHandler);
            this.moveStart(pos);
        },

        _mouseUp: function (pos, evt) {
            this.moveEnd();
            stopListen('mousemove', this._mouseMoveHandler);
            stopListen('mouseup', this._mouseUpHandler);
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
