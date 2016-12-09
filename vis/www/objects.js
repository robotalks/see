(function (exports) {
    'use strict';

    var ObjectBase = Class({
        constructor: function (props, world) {
            this.properties = props;
            this.world = world;
        },

        id: function () {
            return this.properties.id;
        },

        type: function () {
            return this.properties.type;
        },

        render: function (elem, outer) {
            this.outerElem = elem;
            this.outer = outer;
            if (typeof(this.createContent) == 'function') {
                var child = this.createContent(elem, outer);
                if (child != null) {
                    this.setElem(child);
                }
            }
        },

        measure: function () {
            return vis.measure(this.properties);
        },

        applyStyles: function (elem) {
            elem.classList.add('object');
            elem.classList.add(this.properties.type);
            if (this.properties.style != null) {
                elem.classList.add(this.properties.style);
            }
            if (Array.isArray(this.properties.styles)) {
                this.properties.styles.forEach(function (style) {
                    elem.classList.add(style);
                });
            }
            return this;
        },

        setElem: function (elem) {
            this.applyStyles(elem);
            this.outerElem.appendChild(elem);
            return this;
        },

        emit: function (msg) {
            if (!Array.isArray(msg)) {
                msg = [msg];
            }
            this.world.emit(msg);
        }
    });

    var CanvasObject = Class(ObjectBase, {
        constructor: function () {
            ObjectBase.prototype.constructor.apply(this, arguments);
        },

        createContent: function () {
            this._canvas = document.createElement('canvas');
            return this._canvas;
        },

        place: function (viewRc) {
            this._canvas.setAttribute('width', viewRc.w);
            this._canvas.setAttribute('height', viewRc.h);
            this.paint(viewRc, this._canvas);
        },

        paint: function () {
            // to be overridden
        }
    });

    function defineObject (type, prototype, baseClass) {
        if (baseClass == null) {
            baseClass = ObjectBase;
        }
        var ctor = prototype.constructor;
        prototype.constructor = function () {
            baseClass.prototype.constructor.apply(this, arguments);
            if (ctor != null) {
                ctor.apply(this.arguments);
            }
        };
        var cls = Class(baseClass, prototype, { statics: { Type: type } });
        vis.registerClass(cls);
        return cls;
    }

    function defineCanvasObject (type, prototype) {
        return defineObject(type, prototype, CanvasObject);
    }

    vis.ObjectBase = ObjectBase;
    vis.CanvasObject = CanvasObject;
    vis.defineObject = defineObject;
    vis.defineCanvasObject = defineCanvasObject;
})(window);
