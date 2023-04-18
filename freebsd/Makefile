PORTNAME=       	icapeg
PORTVERSION=    	1.0.0
DISTVERSIONPREFIX=  v
CATEGORIES=     	security

MAINTAINER=     	support@egirna.com
COMMENT=        	An ICAP server implementation
WWW=            	https://github.com/egirna/icapeg

LICENSE=        	APACHE20
LICENSE_FILE=   	${WRKSRC}/LICENSE

USES=           	go:modules
USE_GITHUB=	    	yes
GH_ACCOUNT=    		egirna
GH_TUPLE=			\
		            davecgh:go-spew:v1.1.1:davecgh_go_spew/vendor/github.com/davecgh/go-spew \
		            dutchcoders:go-clamd:b970184f4d9e:dutchcoders_go_clamd/vendor/github.com/dutchcoders/go-clamd \
		            fsnotify:fsnotify:v1.5.1:fsnotify_fsnotify/vendor/github.com/fsnotify/fsnotify \
		            go-ini:ini:v1.63.2:go_ini_ini/vendor/gopkg.in/ini.v1 \
		            go-yaml:yaml:v2.4.0:go_yaml_yaml/vendor/gopkg.in/yaml.v2 \
		            golang:sys:63515b42dcdf:golang_sys/vendor/golang.org/x/sys \
		            golang:text:v0.3.6:golang_text/vendor/golang.org/x/text \
		            h2non:filetype:v1.0.12:h2non_filetype/vendor/github.com/h2non/filetype \
		            hashicorp:hcl:v1.0.0:hashicorp_hcl/vendor/github.com/hashicorp/hcl \
		            magiconair:properties:v1.8.5:magiconair_properties/vendor/github.com/magiconair/properties \
		            mitchellh:mapstructure:v1.4.2:mitchellh_mapstructure/vendor/github.com/mitchellh/mapstructure \
		            pelletier:go-toml:v1.9.4:pelletier_go_toml/vendor/github.com/pelletier/go-toml \
		            pkg:errors:v0.9.1:pkg_errors/vendor/github.com/pkg/errors \
		            spf13:afero:v1.6.0:spf13_afero/vendor/github.com/spf13/afero \
		            spf13:cast:v1.4.1:spf13_cast/vendor/github.com/spf13/cast \
		            spf13:jwalterweatherman:v1.1.0:spf13_jwalterweatherman/vendor/github.com/spf13/jwalterweatherman \
		            spf13:pflag:v1.0.5:spf13_pflag/vendor/github.com/spf13/pflag \
		            spf13:viper:v1.9.0:spf13_viper/vendor/github.com/spf13/viper \
		            subosito:gotenv:v1.2.0:subosito_gotenv/vendor/github.com/subosito/gotenv \
		            uber-go:atomic:v1.10.0:uber_go_atomic/vendor/go.uber.org/atomic \
		            uber-go:multierr:v1.8.0:uber_go_multierr/vendor/go.uber.org/multierr \
		            uber-go:zap:v1.22.0:uber_go_zap/vendor/go.uber.org/zap \
		            xhit:go-str2duration:v2.0.0:xhit_go_str2duration_v2/vendor/github.com/xhit/go-str2duration/v2

post-install:
	${MKDIR} ${STAGEDIR}${ETCDIR}
	${INSTALL_DATA} ${WRKSRC}/config.toml ${STAGEDIR}${ETCDIR}/config.toml.sample
	${MKDIR} ${STAGEDIR}${DATADIR}
	${INSTALL_DATA} ${WRKSRC}/temp/exception-page.html ${STAGEDIR}${DATADIR}/exception-page.html

PLIST_FILES=    	bin/icapeg \
					"@sample ${ETCDIR}/config.toml.sample" \
					"${DATADIR}/exception-page.html"

.include <bsd.port.mk>
