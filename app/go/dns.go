package main

import (
	"fmt"
	"slices"
	"sync"

	"github.com/miekg/dns"
)

func NewRR(s string) dns.RR { r, _ := dns.NewRR(s); return r }

var subdomains = []string{
	"u.isucon.dev.",
	"ns1.u.isucon.dev.",
	"pipe.u.isucon.dev.",
	"test001.u.isucon.dev.",

	"ayamazaki0.u.isucon.dev.",
	"yoshidamiki0.u.isucon.dev.",
	"hidekimurakami0.u.isucon.dev.",
	"akemikobayashi0.u.isucon.dev.",
	"eishikawa0.u.isucon.dev.",
	"tomoya100.u.isucon.dev.",
	"kobayashiminoru0.u.isucon.dev.",
	"yamadakaori0.u.isucon.dev.",
	"tanakahanako0.u.isucon.dev.",
	"taro660.u.isucon.dev.",
	"hashimotokenichi0.u.isucon.dev.",
	"yuta330.u.isucon.dev.",
	"kenichinakamura0.u.isucon.dev.",
	"taichikato0.u.isucon.dev.",
	"wfujita0.u.isucon.dev.",
	"saitotakuma0.u.isucon.dev.",
	"maayafujiwara0.u.isucon.dev.",
	"akira680.u.isucon.dev.",
	"vmaeda0.u.isucon.dev.",
	"jnakamura0.u.isucon.dev.",
	"suzukitsubasa0.u.isucon.dev.",
	"yoshidatomoya0.u.isucon.dev.",
	"qendo0.u.isucon.dev.",
	"haruka030.u.isucon.dev.",
	"saitotakuma1.u.isucon.dev.",
	"bsuzuki0.u.isucon.dev.",
	"shohei720.u.isucon.dev.",
	"naoko980.u.isucon.dev.",
	"suzukiryohei0.u.isucon.dev.",
	"kobayashisayuri0.u.isucon.dev.",
	"ykobayashi0.u.isucon.dev.",
	"asuzuki0.u.isucon.dev.",
	"sotaro880.u.isucon.dev.",
	"nyamaguchi0.u.isucon.dev.",
	"momokohashimoto0.u.isucon.dev.",
	"suzukinaoki0.u.isucon.dev.",
	"wgoto0.u.isucon.dev.",
	"tomoya540.u.isucon.dev.",
	"fujitayoichi0.u.isucon.dev.",
	"kaorikato0.u.isucon.dev.",
	"chiyo810.u.isucon.dev.",
	"yito0.u.isucon.dev.",
	"tomoyakato0.u.isucon.dev.",
	"ryosukeabe0.u.isucon.dev.",
	"smatsumoto0.u.isucon.dev.",
	"vwatanabe0.u.isucon.dev.",
	"sayuri650.u.isucon.dev.",
	"takahashinaoto0.u.isucon.dev.",
	"hiroshisuzuki0.u.isucon.dev.",
	"xyamamoto0.u.isucon.dev.",
	"osaito0.u.isucon.dev.",
	"esato0.u.isucon.dev.",
	"oshimizu0.u.isucon.dev.",
	"yamadamituru0.u.isucon.dev.",
	"wmori0.u.isucon.dev.",
	"saitorei0.u.isucon.dev.",
	"kimuramiki0.u.isucon.dev.",
	"sasakiyosuke0.u.isucon.dev.",
	"kumiko410.u.isucon.dev.",
	"qkobayashi0.u.isucon.dev.",
	"akondo0.u.isucon.dev.",
	"ywatanabe0.u.isucon.dev.",
	"otanaoki0.u.isucon.dev.",
	"naotosasaki0.u.isucon.dev.",
	"momokosuzuki0.u.isucon.dev.",
	"maayayoshida0.u.isucon.dev.",
	"hashimotoasuka0.u.isucon.dev.",
	"maayanakagawa0.u.isucon.dev.",
	"ltanaka0.u.isucon.dev.",
	"jyamada0.u.isucon.dev.",
	"yasuhiro650.u.isucon.dev.",
	"hiroshitanaka0.u.isucon.dev.",
	"yamashitaakemi0.u.isucon.dev.",
	"hidekiishikawa0.u.isucon.dev.",
	"kanatakahashi0.u.isucon.dev.",
	"akemiito0.u.isucon.dev.",
	"manabu800.u.isucon.dev.",
	"shota790.u.isucon.dev.",
	"naokoaoki0.u.isucon.dev.",
	"gyamamoto0.u.isucon.dev.",
	"xyamaguchi0.u.isucon.dev.",
	"katokumiko0.u.isucon.dev.",
	"xyamada0.u.isucon.dev.",
	"qyamashita0.u.isucon.dev.",
	"pwatanabe0.u.isucon.dev.",
	"ryoheiishikawa0.u.isucon.dev.",
	"nakamuraharuka0.u.isucon.dev.",
	"pfukuda0.u.isucon.dev.",
	"saitoyosuke0.u.isucon.dev.",
	"morijun0.u.isucon.dev.",
	"hiroshiokamoto0.u.isucon.dev.",
	"kazuyahayashi0.u.isucon.dev.",
	"anakajima0.u.isucon.dev.",
	"maaya110.u.isucon.dev.",
	"kazuya250.u.isucon.dev.",
	"kondoakira0.u.isucon.dev.",
	"atsushi920.u.isucon.dev.",
	"hanakokondo0.u.isucon.dev.",
	"naoki220.u.isucon.dev.",
	"hiroshi180.u.isucon.dev.",
	"jun820.u.isucon.dev.",
	"onakamura0.u.isucon.dev.",
	"takuma040.u.isucon.dev.",
	"kobayashiyasuhiro0.u.isucon.dev.",
	"reiwatanabe0.u.isucon.dev.",
	"hideki630.u.isucon.dev.",
	"yutayoshida0.u.isucon.dev.",
	"yutamori0.u.isucon.dev.",
	"ryosukeyamamoto0.u.isucon.dev.",
	"manabuota0.u.isucon.dev.",
	"nsakamoto0.u.isucon.dev.",
	"wtanaka0.u.isucon.dev.",
	"inakamura0.u.isucon.dev.",
	"ymori0.u.isucon.dev.",
	"nfujii0.u.isucon.dev.",
	"osamu950.u.isucon.dev.",
	"yuinakagawa0.u.isucon.dev.",
	"uhayashi0.u.isucon.dev.",
	"momoko070.u.isucon.dev.",
	"myamada0.u.isucon.dev.",
	"zyoshida0.u.isucon.dev.",
	"watanabesatomi0.u.isucon.dev.",
	"watanabemanabu0.u.isucon.dev.",
	"momokosuzuki1.u.isucon.dev.",
	"jokada0.u.isucon.dev.",
	"hanako640.u.isucon.dev.",
	"rmatsumoto0.u.isucon.dev.",
	"maayasasaki0.u.isucon.dev.",
	"ryohei450.u.isucon.dev.",
	"takuma570.u.isucon.dev.",
	"yokotanaka0.u.isucon.dev.",
	"mai270.u.isucon.dev.",
	"rikasakamoto0.u.isucon.dev.",
	"taro150.u.isucon.dev.",
	"suzukimomoko0.u.isucon.dev.",
	"watanabeyoichi0.u.isucon.dev.",
	"yuki620.u.isucon.dev.",
	"tsubasainoue0.u.isucon.dev.",
	"yokotakahashi0.u.isucon.dev.",
	"yumikohayashi0.u.isucon.dev.",
	"yamaguchiyumiko0.u.isucon.dev.",
	"kkato0.u.isucon.dev.",
	"minoruyoshida0.u.isucon.dev.",
	"tmurakami0.u.isucon.dev.",
	"hasegawanaoto0.u.isucon.dev.",
	"kumikowatanabe0.u.isucon.dev.",
	"lota0.u.isucon.dev.",
	"yoichi440.u.isucon.dev.",
	"sayurikondo0.u.isucon.dev.",
	"xogawa0.u.isucon.dev.",
	"naoki250.u.isucon.dev.",
	"eokada0.u.isucon.dev.",
	"satomiyamamoto0.u.isucon.dev.",
	"asukamaeda0.u.isucon.dev.",
	"momokowatanabe0.u.isucon.dev.",
	"nanamisuzuki0.u.isucon.dev.",
	"gotomanabu0.u.isucon.dev.",
	"maiyamazaki0.u.isucon.dev.",
	"junito0.u.isucon.dev.",
	"aokikazuya0.u.isucon.dev.",
	"shohei040.u.isucon.dev.",
	"momokotanaka0.u.isucon.dev.",
	"atsushimatsumoto0.u.isucon.dev.",
	"taichifukuda0.u.isucon.dev.",
	"haruka630.u.isucon.dev.",
	"tanakashohei0.u.isucon.dev.",
	"qnishimura0.u.isucon.dev.",
	"mkondo0.u.isucon.dev.",
	"kimurayui0.u.isucon.dev.",
	"tomoyanakajima0.u.isucon.dev.",
	"naoko310.u.isucon.dev.",
	"yukiokada0.u.isucon.dev.",
	"tsubasasuzuki0.u.isucon.dev.",
	"yutatakahashi0.u.isucon.dev.",
	"kyosuke140.u.isucon.dev.",
	"tomoya190.u.isucon.dev.",
	"ryohei610.u.isucon.dev.",
	"fukudahiroshi0.u.isucon.dev.",
	"tyamaguchi0.u.isucon.dev.",
	"rsasaki0.u.isucon.dev.",
	"satoyoichi0.u.isucon.dev.",
	"rikawatanabe0.u.isucon.dev.",
	"vokada0.u.isucon.dev.",
	"mituru070.u.isucon.dev.",
	"nakajimayui0.u.isucon.dev.",
	"nanamiota0.u.isucon.dev.",
	"smatsuda0.u.isucon.dev.",
	"suzukijun0.u.isucon.dev.",
	"saitosayuri0.u.isucon.dev.",
	"haruka730.u.isucon.dev.",
	"yamamotoakemi0.u.isucon.dev.",
	"satoyosuke0.u.isucon.dev.",
	"maayasato0.u.isucon.dev.",
	"nanami830.u.isucon.dev.",
	"kanaokamoto0.u.isucon.dev.",
	"usuzuki0.u.isucon.dev.",
	"manabusasaki0.u.isucon.dev.",
	"kenichi870.u.isucon.dev.",
	"tomoyasato0.u.isucon.dev.",
	"mituruhayashi0.u.isucon.dev.",
	"junnishimura0.u.isucon.dev.",
	"fujiwaramituru0.u.isucon.dev.",
	"yoshidamai0.u.isucon.dev.",
	"yosuke350.u.isucon.dev.",
	"xwatanabe0.u.isucon.dev.",
	"naokimatsumoto0.u.isucon.dev.",
	"kenichi170.u.isucon.dev.",
	"naoko340.u.isucon.dev.",
	"nakajimasotaro0.u.isucon.dev.",
	"eito0.u.isucon.dev.",
	"kenichi470.u.isucon.dev.",
	"qabe0.u.isucon.dev.",
	"junkondo0.u.isucon.dev.",
	"yosukewatanabe0.u.isucon.dev.",
	"ryosuke040.u.isucon.dev.",
	"itosotaro0.u.isucon.dev.",
	"kimuramikako0.u.isucon.dev.",
	"reiogawa0.u.isucon.dev.",
	"myamada1.u.isucon.dev.",
	"yamazakisayuri0.u.isucon.dev.",
	"gotoakemi0.u.isucon.dev.",
	"shohei660.u.isucon.dev.",
	"phasegawa0.u.isucon.dev.",
	"kana990.u.isucon.dev.",
	"yosuke710.u.isucon.dev.",
	"rika420.u.isucon.dev.",
	"ikedajun0.u.isucon.dev.",
	"takuma250.u.isucon.dev.",
	"yoichihayashi0.u.isucon.dev.",
	"yoichi441.u.isucon.dev.",
	"zmiura0.u.isucon.dev.",
	"ftakahashi0.u.isucon.dev.",
	"oinoue0.u.isucon.dev.",
	"osamu920.u.isucon.dev.",
	"takahashikumiko0.u.isucon.dev.",
	"tsubasamaeda0.u.isucon.dev.",
	"ryosuke220.u.isucon.dev.",
	"yoichimurakami0.u.isucon.dev.",
	"abeyumiko0.u.isucon.dev.",
	"tishii0.u.isucon.dev.",
	"vkato0.u.isucon.dev.",
	"mituru710.u.isucon.dev.",
	"asukamurakami0.u.isucon.dev.",
	"yuiwatanabe0.u.isucon.dev.",
	"jokamoto0.u.isucon.dev.",
	"akirafujiwara0.u.isucon.dev.",
	"lgoto0.u.isucon.dev.",
	"akiratakahashi0.u.isucon.dev.",
	"lhashimoto0.u.isucon.dev.",
	"ogawarika0.u.isucon.dev.",
	"manabu620.u.isucon.dev.",
	"osamuyamaguchi0.u.isucon.dev.",
	"fujiwararei0.u.isucon.dev.",
	"yosukewatanabe1.u.isucon.dev.",
	"asuka500.u.isucon.dev.",
	"otamaaya0.u.isucon.dev.",
	"okamotonaoko0.u.isucon.dev.",
	"tokamoto0.u.isucon.dev.",
	"saitomituru0.u.isucon.dev.",
	"suzukitakuma0.u.isucon.dev.",
	"suzukimaaya0.u.isucon.dev.",
	"nanamitanaka0.u.isucon.dev.",
	"suzukikana0.u.isucon.dev.",
	"sakamotoshohei0.u.isucon.dev.",
	"kanayamamoto0.u.isucon.dev.",
	"xtakahashi0.u.isucon.dev.",
	"wsuzuki0.u.isucon.dev.",
	"taichi990.u.isucon.dev.",
	"saitoyuki0.u.isucon.dev.",
	"rei050.u.isucon.dev.",
	"nmaeda0.u.isucon.dev.",
	"csuzuki0.u.isucon.dev.",
	"takahashiatsushi0.u.isucon.dev.",
	"yukifukuda0.u.isucon.dev.",
	"yumiko680.u.isucon.dev.",
	"yamadayoko0.u.isucon.dev.",
	"abeosamu0.u.isucon.dev.",
	"taro210.u.isucon.dev.",
	"katomai0.u.isucon.dev.",
	"hasegawakumiko0.u.isucon.dev.",
	"vhashimoto0.u.isucon.dev.",
	"tanakamai0.u.isucon.dev.",
	"hmori0.u.isucon.dev.",
	"naokoyamaguchi0.u.isucon.dev.",
	"yamaguchiyuta0.u.isucon.dev.",
	"iyamamoto0.u.isucon.dev.",
	"yutanakajima0.u.isucon.dev.",
	"kyosukesasaki0.u.isucon.dev.",
	"satosatomi0.u.isucon.dev.",
	"kkato1.u.isucon.dev.",
	"kumiko980.u.isucon.dev.",
	"jishii0.u.isucon.dev.",
	"yutafujiwara0.u.isucon.dev.",
	"fukudamomoko0.u.isucon.dev.",
	"kumikokobayashi0.u.isucon.dev.",
	"naoki350.u.isucon.dev.",
	"sayuri230.u.isucon.dev.",
	"morijun1.u.isucon.dev.",
	"mituru840.u.isucon.dev.",
	"eyoshida0.u.isucon.dev.",
	"taichi460.u.isucon.dev.",
	"esasaki0.u.isucon.dev.",
	"yumikonakamura0.u.isucon.dev.",
	"akemi540.u.isucon.dev.",
	"kazuya770.u.isucon.dev.",
	"suzukiharuka0.u.isucon.dev.",
	"harukagoto0.u.isucon.dev.",
	"vkato1.u.isucon.dev.",
	"bmatsumoto0.u.isucon.dev.",
	"yamazakikaori0.u.isucon.dev.",
	"yamashitasatomi0.u.isucon.dev.",
	"fujiimikako0.u.isucon.dev.",
	"ekato0.u.isucon.dev.",
	"asukaito0.u.isucon.dev.",
	"aokimiki0.u.isucon.dev.",
	"katotaichi0.u.isucon.dev.",
	"harukasasaki0.u.isucon.dev.",
	"manabuyamamoto0.u.isucon.dev.",
	"etanaka0.u.isucon.dev.",
	"nakagawamituru0.u.isucon.dev.",
	"hiroshikobayashi0.u.isucon.dev.",
	"satomi900.u.isucon.dev.",
	"akira180.u.isucon.dev.",
	"manabu710.u.isucon.dev.",
	"minoru160.u.isucon.dev.",
	"fujiwarataro0.u.isucon.dev.",
	"tomoya240.u.isucon.dev.",
	"taichinakajima0.u.isucon.dev.",
	"atsushitakahashi0.u.isucon.dev.",
	"yoichi150.u.isucon.dev.",
	"fito0.u.isucon.dev.",
	"satomitakahashi0.u.isucon.dev.",
	"akemi230.u.isucon.dev.",
	"akato0.u.isucon.dev.",
	"asuka250.u.isucon.dev.",
	"atsushiinoue0.u.isucon.dev.",
	"sasakirei0.u.isucon.dev.",
	"yamadamiki0.u.isucon.dev.",
	"taro890.u.isucon.dev.",
	"yamamotohiroshi0.u.isucon.dev.",
	"vtakahashi0.u.isucon.dev.",
	"yoichiyoshida0.u.isucon.dev.",
	"nishimurashohei0.u.isucon.dev.",
	"zsasaki0.u.isucon.dev.",
	"shimizumaaya0.u.isucon.dev.",
	"inoueosamu0.u.isucon.dev.",
	"hsato0.u.isucon.dev.",
	"kimuraharuka0.u.isucon.dev.",
	"takumayoshida0.u.isucon.dev.",
	"ikedahideki0.u.isucon.dev.",
	"kobayashitsubasa0.u.isucon.dev.",
	"jmurakami0.u.isucon.dev.",
	"kaori320.u.isucon.dev.",
	"suzukimiki0.u.isucon.dev.",
	"gtanaka0.u.isucon.dev.",
	"shota220.u.isucon.dev.",
	"fyamada0.u.isucon.dev.",
	"nishimuraminoru0.u.isucon.dev.",
	"ttanaka0.u.isucon.dev.",
	"satomi280.u.isucon.dev.",
	"maaya390.u.isucon.dev.",
	"snakamura0.u.isucon.dev.",
	"okamototsubasa0.u.isucon.dev.",
	"etakahashi0.u.isucon.dev.",
	"taro500.u.isucon.dev.",
	"yasuhirotakahashi0.u.isucon.dev.",
	"uhasegawa0.u.isucon.dev.",
	"miki470.u.isucon.dev.",
	"ryosukeito0.u.isucon.dev.",
	"kanasaito0.u.isucon.dev.",
	"dyamashita0.u.isucon.dev.",
	"taichi180.u.isucon.dev.",
	"iyamazaki0.u.isucon.dev.",
	"yui950.u.isucon.dev.",
	"kumikomaeda0.u.isucon.dev.",
	"yuiwatanabe1.u.isucon.dev.",
	"suzukishota0.u.isucon.dev.",
	"vsuzuki0.u.isucon.dev.",
	"tanakahiroshi0.u.isucon.dev.",
	"yosukekimura0.u.isucon.dev.",
	"saitoryosuke0.u.isucon.dev.",
	"ymori1.u.isucon.dev.",
	"xnakajima0.u.isucon.dev.",
	"zyamazaki0.u.isucon.dev.",
	"mikakoyamamoto0.u.isucon.dev.",
	"tanakahideki0.u.isucon.dev.",
	"ekato1.u.isucon.dev.",
	"momokomori0.u.isucon.dev.",
	"tsubasa260.u.isucon.dev.",
	"mikako290.u.isucon.dev.",
	"yoko160.u.isucon.dev.",
	"yuki480.u.isucon.dev.",
	"hasegawakumiko1.u.isucon.dev.",
	"itokana0.u.isucon.dev.",
	"fkobayashi0.u.isucon.dev.",
	"tomoya630.u.isucon.dev.",
	"yuki730.u.isucon.dev.",
	"vyamaguchi0.u.isucon.dev.",
	"tshimizu0.u.isucon.dev.",
	"pota0.u.isucon.dev.",
	"hiroshi960.u.isucon.dev.",
	"katoyui0.u.isucon.dev.",
	"mkondo1.u.isucon.dev.",
	"rika500.u.isucon.dev.",
	"einoue0.u.isucon.dev.",
	"yamashitayumiko0.u.isucon.dev.",
	"kazuya680.u.isucon.dev.",
	"asukakobayashi0.u.isucon.dev.",
	"miki320.u.isucon.dev.",
	"yumikosuzuki0.u.isucon.dev.",
	"naotoito0.u.isucon.dev.",
	"esuzuki0.u.isucon.dev.",
	"hiroshiaoki0.u.isucon.dev.",
	"vishikawa0.u.isucon.dev.",
	"akiraito0.u.isucon.dev.",
	"mikako020.u.isucon.dev.",
	"miurasatomi0.u.isucon.dev.",
	"etakahashi1.u.isucon.dev.",
	"tarosato0.u.isucon.dev.",
	"matsumotomanabu0.u.isucon.dev.",
	"naoko560.u.isucon.dev.",
	"kaori570.u.isucon.dev.",
	"minoru210.u.isucon.dev.",
	"jsuzuki0.u.isucon.dev.",
	"tanakashohei1.u.isucon.dev.",
	"tyamada0.u.isucon.dev.",
	"kenichi160.u.isucon.dev.",
	"shotayamada0.u.isucon.dev.",
	"hiroshi070.u.isucon.dev.",
	"hidekisakamoto0.u.isucon.dev.",
	"naokiikeda0.u.isucon.dev.",
	"nakagawamaaya0.u.isucon.dev.",
	"hiroshitanaka1.u.isucon.dev.",
	"yamashitamaaya0.u.isucon.dev.",
	"momokoabe0.u.isucon.dev.",
	"vito0.u.isucon.dev.",
	"sakamotomikako0.u.isucon.dev.",
	"asaito0.u.isucon.dev.",
	"ryosukewatanabe0.u.isucon.dev.",
	"nanami140.u.isucon.dev.",
	"tanakahanako1.u.isucon.dev.",
	"rei390.u.isucon.dev.",
	"rito0.u.isucon.dev.",
	"akemiinoue0.u.isucon.dev.",
	"jendo0.u.isucon.dev.",
	"cwatanabe0.u.isucon.dev.",
	"rikakobayashi0.u.isucon.dev.",
	"kaoriokamoto0.u.isucon.dev.",
	"momoko350.u.isucon.dev.",
	"sayuri130.u.isucon.dev.",
	"kaoritanaka0.u.isucon.dev.",
	"mikakokobayashi0.u.isucon.dev.",
	"naoto500.u.isucon.dev.",
	"suzukinaoto0.u.isucon.dev.",
	"myamamoto0.u.isucon.dev.",
	"yuta090.u.isucon.dev.",
	"asato0.u.isucon.dev.",
	"chiyo580.u.isucon.dev.",
	"lmatsumoto0.u.isucon.dev.",
	"kaori190.u.isucon.dev.",
	"mai860.u.isucon.dev.",
	"naokokimura0.u.isucon.dev.",
	"tomoya400.u.isucon.dev.",
	"dota0.u.isucon.dev.",
	"osamu590.u.isucon.dev.",
	"takahashitomoya0.u.isucon.dev.",
	"shimizuakemi0.u.isucon.dev.",
	"endokenichi0.u.isucon.dev.",
	"yoshidaryohei0.u.isucon.dev.",
	"ryosukemurakami0.u.isucon.dev.",
	"reitanaka0.u.isucon.dev.",
	"akemi910.u.isucon.dev.",
	"enakamura0.u.isucon.dev.",
	"ohayashi0.u.isucon.dev.",
	"takumasuzuki0.u.isucon.dev.",
	"suzukijun1.u.isucon.dev.",
	"bsaito0.u.isucon.dev.",
	"maifujiwara0.u.isucon.dev.",
	"ikedayoko0.u.isucon.dev.",
	"kimurakenichi0.u.isucon.dev.",
	"kyosuke210.u.isucon.dev.",
	"shota870.u.isucon.dev.",
	"fyamaguchi0.u.isucon.dev.",
	"miki730.u.isucon.dev.",
	"sayurihashimoto0.u.isucon.dev.",
	"naoki460.u.isucon.dev.",
	"satoyumiko0.u.isucon.dev.",
	"minoruhashimoto0.u.isucon.dev.",
	"nanamimurakami0.u.isucon.dev.",
	"mai600.u.isucon.dev.",
	"taro340.u.isucon.dev.",
	"yosuke950.u.isucon.dev.",
	"rokada0.u.isucon.dev.",
	"nanami710.u.isucon.dev.",
	"zishikawa0.u.isucon.dev.",
	"yutawatanabe0.u.isucon.dev.",
	"suzukiosamu0.u.isucon.dev.",
	"mituru130.u.isucon.dev.",
	"taichi000.u.isucon.dev.",
	"chiyoyamamoto0.u.isucon.dev.",
	"yoshidakyosuke0.u.isucon.dev.",
	"vyamazaki0.u.isucon.dev.",
	"wwatanabe0.u.isucon.dev.",
	"ykimura0.u.isucon.dev.",
	"tanakahideki1.u.isucon.dev.",
	"minoruito0.u.isucon.dev.",
	"shotamatsumoto0.u.isucon.dev.",
	"watanabehiroshi0.u.isucon.dev.",
	"taichi560.u.isucon.dev.",
	"nanami090.u.isucon.dev.",
	"akira800.u.isucon.dev.",
	"suzukiyasuhiro0.u.isucon.dev.",
	"momokofukuda0.u.isucon.dev.",
	"kana350.u.isucon.dev.",
	"tanakanaoto0.u.isucon.dev.",
	"akemi800.u.isucon.dev.",
	"tanakakenichi0.u.isucon.dev.",
	"oyoshida0.u.isucon.dev.",
	"yutamatsumoto0.u.isucon.dev.",
	"hiroshisuzuki1.u.isucon.dev.",
	"yamaguchikazuya0.u.isucon.dev.",
	"yoko800.u.isucon.dev.",
	"lnakamura0.u.isucon.dev.",
	"esato1.u.isucon.dev.",
	"atsushitakahashi1.u.isucon.dev.",
	"eshimizu0.u.isucon.dev.",
	"reitanaka1.u.isucon.dev.",
	"yuki500.u.isucon.dev.",
	"tanakashohei2.u.isucon.dev.",
	"akiranishimura0.u.isucon.dev.",
	"naokiyamamoto0.u.isucon.dev.",
	"shimizuatsushi0.u.isucon.dev.",
	"winoue0.u.isucon.dev.",
	"hiroshifujita0.u.isucon.dev.",
	"mikako380.u.isucon.dev.",
	"gotoshohei0.u.isucon.dev.",
	"yumikoyamamoto0.u.isucon.dev.",
	"gmatsuda0.u.isucon.dev.",
	"kumikokobayashi1.u.isucon.dev.",
	"tsubasatakahashi0.u.isucon.dev.",
	"yyamamoto0.u.isucon.dev.",
	"nakamurajun0.u.isucon.dev.",
	"yasuhirokato0.u.isucon.dev.",
	"vsuzuki1.u.isucon.dev.",
	"kobayashihideki0.u.isucon.dev.",
	"murakamiosamu0.u.isucon.dev.",
	"mikako340.u.isucon.dev.",
	"takumaokada0.u.isucon.dev.",
	"osamusato0.u.isucon.dev.",
	"aogawa0.u.isucon.dev.",
	"ikedayosuke0.u.isucon.dev.",
	"sakamotoyoichi0.u.isucon.dev.",
	"yamadakaori1.u.isucon.dev.",
	"ykobayashi1.u.isucon.dev.",
	"fukudananami0.u.isucon.dev.",
	"yasuhirohashimoto0.u.isucon.dev.",
	"akemi670.u.isucon.dev.",
	"rhasegawa0.u.isucon.dev.",
	"naoto740.u.isucon.dev.",
	"satoasuka0.u.isucon.dev.",
	"lyamashita0.u.isucon.dev.",
	"kumiko030.u.isucon.dev.",
	"qikeda0.u.isucon.dev.",
	"ptanaka0.u.isucon.dev.",
	"fmatsuda0.u.isucon.dev.",
	"akiramaeda0.u.isucon.dev.",
	"yoichiyamamoto0.u.isucon.dev.",
	"gsuzuki0.u.isucon.dev.",
	"naokiikeda1.u.isucon.dev.",
	"abekenichi0.u.isucon.dev.",
	"maiyamamoto0.u.isucon.dev.",
	"shimizukumiko0.u.isucon.dev.",
	"zsato0.u.isucon.dev.",
	"manabu050.u.isucon.dev.",
	"shoheinishimura0.u.isucon.dev.",
	"gyoshida0.u.isucon.dev.",
	"chiyo790.u.isucon.dev.",
	"hanakookada0.u.isucon.dev.",
	"xota0.u.isucon.dev.",
	"manabusakamoto0.u.isucon.dev.",
	"jsato0.u.isucon.dev.",
	"kenichi840.u.isucon.dev.",
	"kobayashishota0.u.isucon.dev.",
	"ainoue0.u.isucon.dev.",
	"takuma380.u.isucon.dev.",
	"dikeda0.u.isucon.dev.",
	"dtanaka0.u.isucon.dev.",
	"miturutakahashi0.u.isucon.dev.",
	"fmurakami0.u.isucon.dev.",
	"ishiiakira0.u.isucon.dev.",
	"maayaokamoto0.u.isucon.dev.",
	"asaito1.u.isucon.dev.",
	"atsushi470.u.isucon.dev.",
	"ekobayashi0.u.isucon.dev.",
	"ukobayashi0.u.isucon.dev.",
	"asaito2.u.isucon.dev.",
	"hayashirei0.u.isucon.dev.",
	"suzukisatomi0.u.isucon.dev.",
	"yumiko100.u.isucon.dev.",
	"okadahanako0.u.isucon.dev.",
	"mtanaka0.u.isucon.dev.",
	"hidekiyamamoto0.u.isucon.dev.",
	"yamamotomituru0.u.isucon.dev.",
	"kimuratomoya0.u.isucon.dev.",
	"mikako780.u.isucon.dev.",
	"pgoto0.u.isucon.dev.",
	"harukawatanabe0.u.isucon.dev.",
	"zkimura0.u.isucon.dev.",
	"taichikimura0.u.isucon.dev.",
	"moriminoru0.u.isucon.dev.",
	"thayashi0.u.isucon.dev.",
	"akira740.u.isucon.dev.",
	"suzukiharuka1.u.isucon.dev.",
	"yuta030.u.isucon.dev.",
	"fyamamoto0.u.isucon.dev.",
	"watanaberyosuke0.u.isucon.dev.",
	"maayamatsumoto0.u.isucon.dev.",
	"yoichiendo0.u.isucon.dev.",
	"satokaori0.u.isucon.dev.",
	"gmatsumoto0.u.isucon.dev.",
	"rika250.u.isucon.dev.",
	"jsasaki0.u.isucon.dev.",
	"yutasaito0.u.isucon.dev.",
	"kazuya810.u.isucon.dev.",
	"manabukobayashi0.u.isucon.dev.",
	"hashimotomikako0.u.isucon.dev.",
	"fujiwarahideki0.u.isucon.dev.",
	"yoichi730.u.isucon.dev.",
	"takahashiryosuke0.u.isucon.dev.",
	"tyamaguchi1.u.isucon.dev.",
	"otaminoru0.u.isucon.dev.",
	"momoko990.u.isucon.dev.",
	"kazuyahasegawa0.u.isucon.dev.",
	"mai770.u.isucon.dev.",
	"aishii0.u.isucon.dev.",
	"pokada0.u.isucon.dev.",
	"cyamaguchi0.u.isucon.dev.",
	"akemi500.u.isucon.dev.",
	"jtanaka0.u.isucon.dev.",
	"tnakagawa0.u.isucon.dev.",
	"kondoyoko0.u.isucon.dev.",
	"yoichiogawa0.u.isucon.dev.",
	"hayashikenichi0.u.isucon.dev.",
	"nanami990.u.isucon.dev.",
	"hiroshiendo0.u.isucon.dev.",
	"akirasuzuki0.u.isucon.dev.",
	"maaya140.u.isucon.dev.",
	"itomomoko0.u.isucon.dev.",
	"satorika0.u.isucon.dev.",
	"yoko110.u.isucon.dev.",
	"wyamazaki0.u.isucon.dev.",
	"yoko111.u.isucon.dev.",
	"osato0.u.isucon.dev.",
	"aokitaro0.u.isucon.dev.",
	"gnakamura0.u.isucon.dev.",
	"kobayashiyoko0.u.isucon.dev.",
	"suzukinanami0.u.isucon.dev.",
	"yosukeishikawa0.u.isucon.dev.",
	"manabu450.u.isucon.dev.",
	"naoki630.u.isucon.dev.",
	"kimuranaoki0.u.isucon.dev.",
	"kumiko770.u.isucon.dev.",
	"moriyuta0.u.isucon.dev.",
	"tsubasa300.u.isucon.dev.",
	"hidekinakamura0.u.isucon.dev.",
	"rei080.u.isucon.dev.",
	"shoheiyoshida0.u.isucon.dev.",
	"omatsumoto0.u.isucon.dev.",
	"haruka260.u.isucon.dev.",
	"takuma790.u.isucon.dev.",
	"satoryosuke0.u.isucon.dev.",
	"maedamaaya0.u.isucon.dev.",
	"mikakosasaki0.u.isucon.dev.",
	"vmiura0.u.isucon.dev.",
	"rika570.u.isucon.dev.",
	"myamaguchi0.u.isucon.dev.",
	"minoru110.u.isucon.dev.",
	"ishikawakana0.u.isucon.dev.",
	"yoko300.u.isucon.dev.",
	"tsuzuki0.u.isucon.dev.",
	"kenichiwatanabe0.u.isucon.dev.",
	"naokosakamoto0.u.isucon.dev.",
	"fujitataichi0.u.isucon.dev.",
	"naokigoto0.u.isucon.dev.",
	"yuta210.u.isucon.dev.",
	"sgoto0.u.isucon.dev.",
	"momokoishii0.u.isucon.dev.",
	"yoshidahideki0.u.isucon.dev.",
	"takumainoue0.u.isucon.dev.",
	"suzukiyuta0.u.isucon.dev.",
	"manabu370.u.isucon.dev.",
	"asuka120.u.isucon.dev.",
	"manabu621.u.isucon.dev.",
	"gogawa0.u.isucon.dev.",
	"yamashitaosamu0.u.isucon.dev.",
	"jnakamura1.u.isucon.dev.",
	"atakahashi0.u.isucon.dev.",
	"osamukobayashi0.u.isucon.dev.",
	"satomisato0.u.isucon.dev.",
	"satorika1.u.isucon.dev.",
	"tanakahiroshi1.u.isucon.dev.",
	"satoyuta0.u.isucon.dev.",
	"mai710.u.isucon.dev.",
	"nakamurakaori0.u.isucon.dev.",
	"naoko440.u.isucon.dev.",
	"jokamoto1.u.isucon.dev.",
	"kobayashirei0.u.isucon.dev.",
	"kyosukewatanabe0.u.isucon.dev.",
	"yamaguchijun0.u.isucon.dev.",
	"yoshidahiroshi0.u.isucon.dev.",
	"miurahanako0.u.isucon.dev.",
	"tsubasa790.u.isucon.dev.",
	"hiroshitakahashi0.u.isucon.dev.",
	"yukishimizu0.u.isucon.dev.",
	"inouenaoki0.u.isucon.dev.",
	"yuinakamura0.u.isucon.dev.",
	"hashimotoakemi0.u.isucon.dev.",
	"jwatanabe0.u.isucon.dev.",
	"yosuke890.u.isucon.dev.",
	"mikisato0.u.isucon.dev.",
	"naotoinoue0.u.isucon.dev.",
	"vinoue0.u.isucon.dev.",
	"takuma170.u.isucon.dev.",
	"yamashitatsubasa0.u.isucon.dev.",
	"mituru180.u.isucon.dev.",
	"ttanaka1.u.isucon.dev.",
	"ogawataichi0.u.isucon.dev.",
	"rsakamoto0.u.isucon.dev.",
	"satoakemi0.u.isucon.dev.",
	"lito0.u.isucon.dev.",
	"suzukimaaya1.u.isucon.dev.",
	"nakamurarei0.u.isucon.dev.",
	"momokofujiwara0.u.isucon.dev.",
	"yoko310.u.isucon.dev.",
	"fujiwaramanabu0.u.isucon.dev.",
	"takahashiosamu0.u.isucon.dev.",
	"nakamurakazuya0.u.isucon.dev.",
	"nnakamura0.u.isucon.dev.",
	"akirasuzuki1.u.isucon.dev.",
	"watanabenanami0.u.isucon.dev.",
	"rikaito0.u.isucon.dev.",
	"juntanaka0.u.isucon.dev.",
	"tanakamiki0.u.isucon.dev.",
	"ekato2.u.isucon.dev.",
	"xnishimura0.u.isucon.dev.",
	"ihayashi0.u.isucon.dev.",
	"atsushi921.u.isucon.dev.",
	"mikakoishikawa0.u.isucon.dev.",
	"junmatsuda0.u.isucon.dev.",
	"abejun0.u.isucon.dev.",
	"katoshota0.u.isucon.dev.",
	"shotaishikawa0.u.isucon.dev.",
	"yukikimura0.u.isucon.dev.",
	"xishii0.u.isucon.dev.",
	"rei240.u.isucon.dev.",
	"yutakobayashi0.u.isucon.dev.",
	"kazuya380.u.isucon.dev.",
	"nanamikobayashi0.u.isucon.dev.",
	"zyamada0.u.isucon.dev.",
	"inouehanako0.u.isucon.dev.",
	"yutaota0.u.isucon.dev.",
	"satomikato0.u.isucon.dev.",
	"yoshidayasuhiro0.u.isucon.dev.",
	"akemiikeda0.u.isucon.dev.",
	"yoichikobayashi0.u.isucon.dev.",
	"msuzuki0.u.isucon.dev.",
	"ssato0.u.isucon.dev.",
	"qnakamura0.u.isucon.dev.",
	"ryoheikobayashi0.u.isucon.dev.",
	"naoki880.u.isucon.dev.",
	"sakamotosotaro0.u.isucon.dev.",
	"haruka350.u.isucon.dev.",
	"satonaoki0.u.isucon.dev.",
	"jun430.u.isucon.dev.",
	"sasakimai0.u.isucon.dev.",
	"naoko880.u.isucon.dev.",
	"rikakobayashi1.u.isucon.dev.",
	"nanamikondo0.u.isucon.dev.",
	"yoshidamai1.u.isucon.dev.",
	"nakajimarei0.u.isucon.dev.",
	"morimanabu0.u.isucon.dev.",
	"tanakasayuri0.u.isucon.dev.",
	"momoko040.u.isucon.dev.",
	"kaori580.u.isucon.dev.",
	"haruka100.u.isucon.dev.",
	"akemi600.u.isucon.dev.",
	"miki170.u.isucon.dev.",
	"okamotokyosuke0.u.isucon.dev.",
	"kondonaoto0.u.isucon.dev.",
	"tsubasa610.u.isucon.dev.",
	"ryoheiyamada0.u.isucon.dev.",
	"naoki090.u.isucon.dev.",
	"yamazakimikako0.u.isucon.dev.",
	"minorusato0.u.isucon.dev.",
	"chiyofujii0.u.isucon.dev.",
	"takuma300.u.isucon.dev.",
	"taroyamamoto0.u.isucon.dev.",
	"suzukichiyo0.u.isucon.dev.",
	"wishikawa0.u.isucon.dev.",
	"reinakamura0.u.isucon.dev.",
	"ayamazaki1.u.isucon.dev.",
	"cogawa0.u.isucon.dev.",
	"hasegawamituru0.u.isucon.dev.",
	"cyamaguchi1.u.isucon.dev.",
	"hasegawamomoko0.u.isucon.dev.",
	"suzukishohei0.u.isucon.dev.",
	"qyamazaki0.u.isucon.dev.",
	"kyosuke370.u.isucon.dev.",
	"mgoto0.u.isucon.dev.",
	"yasuhiroogawa0.u.isucon.dev.",
	"maaya250.u.isucon.dev.",
	"kimurayumiko0.u.isucon.dev.",
	"yoko460.u.isucon.dev.",
	"atsushiyamaguchi0.u.isucon.dev.",
	"yamadaasuka0.u.isucon.dev.",
	"wyamaguchi0.u.isucon.dev.",
	"katoyuta0.u.isucon.dev.",
	"yoichi660.u.isucon.dev.",
	"lkato0.u.isucon.dev.",
	"tomoyatakahashi0.u.isucon.dev.",
	"rika370.u.isucon.dev.",
	"inouenaoto0.u.isucon.dev.",
	"kondoshota0.u.isucon.dev.",
	"taro820.u.isucon.dev.",
	"maisasaki0.u.isucon.dev.",
	"gyoshida1.u.isucon.dev.",
	"manabu770.u.isucon.dev.",
	"matsudayoichi0.u.isucon.dev.",
	"nakajimamiki0.u.isucon.dev.",
	"watanabeyoko0.u.isucon.dev.",
	"vwatanabe1.u.isucon.dev.",
	"yamaguchinaoto0.u.isucon.dev.",
	"saitojun0.u.isucon.dev.",
	"naotomori0.u.isucon.dev.",
	"lito1.u.isucon.dev.",
	"ryoheiikeda0.u.isucon.dev.",
	"yuisasaki0.u.isucon.dev.",
	"yishikawa0.u.isucon.dev.",
	"yamazakimituru0.u.isucon.dev.",
	"yumiko050.u.isucon.dev.",
	"okadasayuri0.u.isucon.dev.",
	"skobayashi0.u.isucon.dev.",
	"rika740.u.isucon.dev.",
	"yoko600.u.isucon.dev.",
	"wsato0.u.isucon.dev.",
	"taichi181.u.isucon.dev.",
	"hiroshi800.u.isucon.dev.",
	"yoko420.u.isucon.dev.",
	"naokokondo0.u.isucon.dev.",
	"sakamotorika0.u.isucon.dev.",
	"hiroshiyoshida0.u.isucon.dev.",
	"manabusato0.u.isucon.dev.",
	"rei340.u.isucon.dev.",
	"jsato1.u.isucon.dev.",
	"satomikako0.u.isucon.dev.",
	"murakamihiroshi0.u.isucon.dev.",
	"asukaishii0.u.isucon.dev.",
	"kana550.u.isucon.dev.",
	"fyamaguchi1.u.isucon.dev.",
	"miki190.u.isucon.dev.",
	"dkimura0.u.isucon.dev.",
	"ynakamura0.u.isucon.dev.",
	"endomomoko0.u.isucon.dev.",
	"itoyumiko0.u.isucon.dev.",
	"yamashitayosuke0.u.isucon.dev.",
	"yokokondo0.u.isucon.dev.",
	"tsubasa470.u.isucon.dev.",
	"matsudanaoko0.u.isucon.dev.",
	"miuraatsushi0.u.isucon.dev.",
	"miurataichi0.u.isucon.dev.",
	"manabusato1.u.isucon.dev.",
	"uhasegawa1.u.isucon.dev.",
	"nanamimaeda0.u.isucon.dev.",
	"katotsubasa0.u.isucon.dev.",
	"eyamazaki0.u.isucon.dev.",
	"wyamazaki1.u.isucon.dev.",
	"junyamada0.u.isucon.dev.",
	"akemigoto0.u.isucon.dev.",
	"uokada0.u.isucon.dev.",
	"nishii0.u.isucon.dev.",
	"takumatanaka0.u.isucon.dev.",
	"itomai0.u.isucon.dev.",
	"hidekimaeda0.u.isucon.dev.",
	"osuzuki0.u.isucon.dev.",
	"harukasuzuki0.u.isucon.dev.",
	"watanabeyuta0.u.isucon.dev.",
	"kobayashikaori0.u.isucon.dev.",
	"usuzuki1.u.isucon.dev.",
	"gotokazuya0.u.isucon.dev.",
	"kondotaichi0.u.isucon.dev.",
	"nakamurasatomi0.u.isucon.dev.",
	"ltakahashi0.u.isucon.dev.",
	"jun240.u.isucon.dev.",
	"sayuri860.u.isucon.dev.",
	"jmurakami1.u.isucon.dev.",
	"ftakahashi1.u.isucon.dev.",
	"nanami660.u.isucon.dev.",
	"otaasuka0.u.isucon.dev.",
	"oito0.u.isucon.dev.",
	"satomi720.u.isucon.dev.",
	"ryosuke680.u.isucon.dev.",
	"minoru370.u.isucon.dev.",
	"matsumotokumiko0.u.isucon.dev.",
	"sishii0.u.isucon.dev.",
	"maedanaoko0.u.isucon.dev.",
	"chiyo370.u.isucon.dev.",
	"sasakitakuma0.u.isucon.dev.",
	"reiikeda0.u.isucon.dev.",
	"chiyo540.u.isucon.dev.",
	"yamamotosayuri0.u.isucon.dev.",
	"jota0.u.isucon.dev.",
	"bsasaki0.u.isucon.dev.",
	"bmori0.u.isucon.dev.",
	"itojun0.u.isucon.dev.",
	"atsushikondo0.u.isucon.dev.",
	"jun320.u.isucon.dev.",
	"osamusasaki0.u.isucon.dev.",
	"suzukimaaya2.u.isucon.dev.",
	"shimizuhideki0.u.isucon.dev.",
	"endotomoya0.u.isucon.dev.",
	"yamaguchimanabu0.u.isucon.dev.",
	"rhashimoto0.u.isucon.dev.",
	"rei590.u.isucon.dev.",
	"yosuke490.u.isucon.dev.",
	"tanakaryohei0.u.isucon.dev.",
	"yukinakamura0.u.isucon.dev.",
	"akira580.u.isucon.dev.",
	"minoru400.u.isucon.dev.",
	"yamaguchimomoko0.u.isucon.dev.",
	"hanako410.u.isucon.dev.",
	"hyamamoto0.u.isucon.dev.",
	"maaya910.u.isucon.dev.",
	"sayurikobayashi0.u.isucon.dev.",
	"matsumotonaoko0.u.isucon.dev.",
	"rei250.u.isucon.dev.",
	"momoko980.u.isucon.dev.",
	"asuka370.u.isucon.dev.",
	"asuka100.u.isucon.dev.",
	"yukiito0.u.isucon.dev.",
	"mai500.u.isucon.dev.",
	"manabutanaka0.u.isucon.dev.",
	"taichinakamura0.u.isucon.dev.",
	"csato0.u.isucon.dev.",
	"mikiyoshida0.u.isucon.dev.",
	"fujiiatsushi0.u.isucon.dev.",
	"hayashinaoko0.u.isucon.dev.",
	"kaorikobayashi0.u.isucon.dev.",
	"lfujii0.u.isucon.dev.",
	"wyamaguchi1.u.isucon.dev.",
	"satomifujiwara0.u.isucon.dev.",
	"gyamashita0.u.isucon.dev.",
	"ryohei850.u.isucon.dev.",
	"dnakamura0.u.isucon.dev.",
	"rika040.u.isucon.dev.",
	"satoyuta1.u.isucon.dev.",
	"ynishimura0.u.isucon.dev.",
	"tsubasa240.u.isucon.dev.",
	"yukigoto0.u.isucon.dev.",
	"satomi200.u.isucon.dev.",
	"tsubasashimizu0.u.isucon.dev.",
	"suzukiyumiko0.u.isucon.dev.",
	"nakamurataichi0.u.isucon.dev.",
	"yumiko070.u.isucon.dev.",
	"shotafujiwara0.u.isucon.dev.",
	"maedataro0.u.isucon.dev.",
	"maifujita0.u.isucon.dev.",
	"fujiwarayuki0.u.isucon.dev.",
	"shohei240.u.isucon.dev.",
	"aokirika0.u.isucon.dev.",
	"hanako580.u.isucon.dev.",
	"rgoto0.u.isucon.dev.",
	"yuta100.u.isucon.dev.",
	"yosukeishikawa1.u.isucon.dev.",
	"dokada0.u.isucon.dev.",
	"kimurahanako0.u.isucon.dev.",
	"wnakagawa0.u.isucon.dev.",
	"yukiyoshida0.u.isucon.dev.",
	"naokitanaka0.u.isucon.dev.",
	"ysaito0.u.isucon.dev.",
	"miturusato0.u.isucon.dev.",
	"naoki870.u.isucon.dev.",
	"chiyo310.u.isucon.dev.",
	"naokitanaka1.u.isucon.dev.",
	"chiyonakamura0.u.isucon.dev.",
	"msaito0.u.isucon.dev.",
	"jsuzuki1.u.isucon.dev.",
	"suzukihanako0.u.isucon.dev.",
	"dokada1.u.isucon.dev.",
	"rei850.u.isucon.dev.",
	"kenichihashimoto0.u.isucon.dev.",
	"fnakamura0.u.isucon.dev.",
	"hasegawatakuma0.u.isucon.dev.",
	"lyamamoto0.u.isucon.dev.",
	"takuma600.u.isucon.dev.",
	"manabutakahashi0.u.isucon.dev.",
	"ryosukesakamoto0.u.isucon.dev.",
	"itoharuka0.u.isucon.dev.",
	"satomi130.u.isucon.dev.",
	"tomoya450.u.isucon.dev.",
}
var muSubdomains = sync.RWMutex{}

func addSubdomain(subdomain string) {
	muSubdomains.Lock()
	defer muSubdomains.Unlock()
	subdomains = append(subdomains, subdomain)
}

func startDNS() error {
	dns.HandleFunc("u.isucon.dev.", func(w dns.ResponseWriter, r *dns.Msg) {
		m := new(dns.Msg)
		m.SetReply(r)
		if r.Question[0].Qtype == dns.TypeNS && r.Question[0].Name == "u.isucon.dev." {
			m.Answer = []dns.RR{
				NewRR("u.isucon.dev. 120 IN NS ns1.u.isucon.dev."),
			}
			m.Extra = []dns.RR{
				NewRR("ns1.u.isucon.dev. 120 IN A 54.178.156.176"),
			}
		} else {
			muSubdomains.RLock()
			defer muSubdomains.RUnlock()

			if slices.Contains(subdomains, r.Question[0].Name) {
				m.Answer = []dns.RR{
					NewRR(r.Question[0].Name + " 120 IN A 54.178.156.176"),
				}
			} else {
				m.Rcode = dns.RcodeNameError
				m.Ns = []dns.RR{
					NewRR("u.isucon.dev.		60	IN	SOA	ns1.u.isucon.dev. hostmaster.u.isucon.dev. 0 10800 3600 604800 3600"),
				}
			}
		}
		w.WriteMsg(m)
	})

	fmt.Println(">>>> STARTING DNS SERVER <<<<")

	srv := &dns.Server{Addr: ":53", Net: "udp"}
	return srv.ListenAndServe()
}
