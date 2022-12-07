package fastrand

import "math"

/*
 * Exponential distribution
 *
 * See "The Ziggurat Method for Generating Random Variables"
 * (Marsaglia & Tsang, 2000)
 * https://www.jstatsoft.org/v05/i08/paper [pdf]
 *
 * Fixed correlation and increased number of distinct results generated,
 * see https://github.com/flyingmutant/rand/issues/3
 */

// ExpFloat64 returns an exponentially distributed float64 in the range
// (0, +math.MaxFloat64] with an exponential distribution whose rate parameter
// (lambda) is 1 and whose mean is 1/lambda (1).
// To produce a distribution with a different rate parameter,
// callers can adjust the output using:
//
//	sample = ExpFloat64() / desiredRateParameter
func ExpFloat64() float64 {
	for {
		v := Uint64()
		j := v >> 11
		i := v & 0xFF
		x := float64(j) * we[i]
		if j < ke[i] {
			return x
		}
		if i == 0 {
			return re - math.Log(Float64())
		}
		if fe[i]+Float64()*(fe[i-1]-fe[i]) < math.Exp(-x) {
			return x
		}
	}
}

// ExpFloat64 returns an exponentially distributed float64 in the range
// (0, +math.MaxFloat64] with an exponential distribution whose rate parameter
// (lambda) is 1 and whose mean is 1/lambda (1).
// To produce a distribution with a different rate parameter,
// callers can adjust the output using:
//
//	sample = ExpFloat64() / desiredRateParameter
func (p *PCG64) ExpFloat64() float64 {
	for {
		v := p.Uint64()
		j := v >> 11
		i := v & 0xFF
		x := float64(j) * we[i]
		if j < ke[i] {
			return x
		}
		if i == 0 {
			return re - math.Log(p.Float64())
		}
		if fe[i]+p.Float64()*(fe[i-1]-fe[i]) < math.Exp(-x) {
			return x
		}
	}
}

const (
	re = 7.69711747013104972
)

var ke = [256]uint64{
	0x1c5214272497c5, 0x0, 0x137d5bd79c3137, 0x186ef58e3f3bf4,
	0x1a9bb7320eb0a2, 0x1bd127f7194473, 0x1c951d0f886514, 0x1d1bfe2d5c3970,
	0x1d7e5bd56b18b1, 0x1dc934dd172c6f, 0x1e0409dfac9dc8, 0x1e337b71d47835,
	0x1e5a8b177cb7a0, 0x1e7b42096f046d, 0x1e970daf08ae3c, 0x1eaef5b14ef09e,
	0x1ec3bd07b46558, 0x1ed5f6f08799cd, 0x1ee614ae6e5688, 0x1ef46eca361cd0,
	0x1f014b76ddd4a3, 0x1f0ce313a796b5, 0x1f176369f1f779, 0x1f20f20c452572,
	0x1f29ae1951a876, 0x1f31b18fb95533, 0x1f39125157c107, 0x1f3fe2eb6e694c,
	0x1f463332d788fd, 0x1f4c10bf1d3a0f, 0x1f51874c5c3323, 0x1f56a109c3ecc1,
	0x1f5b66d9099997, 0x1f5fe08210d08c, 0x1f6414dd445770, 0x1f6809f685967a,
	0x1f6bc52a2b02e7, 0x1f6f4b3d32e4f5, 0x1f72a07190f139, 0x1f75c8974d09da,
	0x1f78c71b045cc0, 0x1f7b9f12413ff5, 0x1f7e5346079f8a, 0x1f80e63be21138,
	0x1f835a3dad9163, 0x1f85b16056b913, 0x1f87ed89b24263, 0x1f8a10759374fb,
	0x1f8c1bba3d39ad, 0x1f8e10cc45d04a, 0x1f8ff102013e17, 0x1f91bd968358e1,
	0x1f9377ac47afda, 0x1f95204f8b64db, 0x1f96b878633893, 0x1f98410c968892,
	0x1f99bae146ba81, 0x1f9b26bc697f00, 0x1f9c85561b717a, 0x1f9dd759cfd804,
	0x1f9f1d6761a1cf, 0x1fa058140936c1, 0x1fa187eb3a333a, 0x1fa2ad6f6bc4fd,
	0x1fa3c91ace0683, 0x1fa4db5fee6aa2, 0x1fa5e4aa4d097d, 0x1fa6e55ee46783,
	0x1fa7dddca51ec5, 0x1fa8ce7ce6a876, 0x1fa9b793ce5ff1, 0x1faa9970adb858,
	0x1fab745e588233, 0x1fac48a3740587, 0x1fad1682bf9feb, 0x1fadde3b5782c0,
	0x1faea008f21d6c, 0x1faf5c2418b07d, 0x1fb012c25b7a13, 0x1fb0c41681dff3,
	0x1fb17050b6f1fc, 0x1fb2179eb2963b, 0x1fb2ba2bdfa84b, 0x1fb358217f4e19,
	0x1fb3f1a6c9be0d, 0x1fb486e10cacd7, 0x1fb517f3c793fe, 0x1fb5a500c5fdaa,
	0x1fb62e2837fe5a, 0x1fb6b388c9010b, 0x1fb7353fb5079a, 0x1fb7b368dc7da9,
	0x1fb82e1ed6ba0a, 0x1fb8a57b0347f6, 0x1fb919959a0f74, 0x1fb98a85ba7204,
	0x1fb9f861796f26, 0x1fba633deee287, 0x1fbacb2f41ec17, 0x1fbb3048b49146,
	0x1fbb929caea4e4, 0x1fbbf23cc8029d, 0x1fbc4f39d22996, 0x1fbca9a3e140d5,
	0x1fbd018a548fa0, 0x1fbd56fbde729c, 0x1fbdaa068bd66c, 0x1fbdfab7cb3f42,
	0x1fbe491c7364de, 0x1fbe9540c96960, 0x1fbedf3086b129, 0x1fbf26f6de6175,
	0x1fbf6c9e828ae3, 0x1fbfb031a904c4, 0x1fbff1ba0ffdb1, 0x1fc03141024589,
	0x1fc06ecf5b54b3, 0x1fc0aa6d8b1428, 0x1fc0e42399698b, 0x1fc11bf9298a65,
	0x1fc151f57d1943, 0x1fc1861f770f4c, 0x1fc1b87d9e74b4, 0x1fc1e91620ea43,
	0x1fc217eed505de, 0x1fc2450d3c8400, 0x1fc27076864fc2, 0x1fc29a2f906310,
	0x1fc2c23ce98046, 0x1fc2e8a2d2c6b5, 0x1fc30d654122ee, 0x1fc33087de9c0e,
	0x1fc3520e0b7ec8, 0x1fc371fadf66f9, 0x1fc390512a2888, 0x1fc3ad137497fa,
	0x1fc3c844013349, 0x1fc3e1e4ccab40, 0x1fc3f9f78e4da8, 0x1fc4107db85062,
	0x1fc4257877fd68, 0x1fc438e8b5bfc7, 0x1fc44acf15112b, 0x1fc45b2bf447e9,
	0x1fc469ff6c4506, 0x1fc477495001b2, 0x1fc483092bfbb9, 0x1fc48d3e457ff7,
	0x1fc495e799d21b, 0x1fc49d03dd30b1, 0x1fc4a29179b434, 0x1fc4a68e8e07fc,
	0x1fc4a8f8ebfb8d, 0x1fc4a9ce16ea9f, 0x1fc4a90b41fa36, 0x1fc4a6ad4e28a1,
	0x1fc4a2b0c82e76, 0x1fc49d11e62de3, 0x1fc495cc852df4, 0x1fc48cdc265ec1,
	0x1fc4823bec237a, 0x1fc475e696dee7, 0x1fc467d6817e83, 0x1fc458059dc038,
	0x1fc4466d702e22, 0x1fc433070bcb9a, 0x1fc41dcb0d6e0e, 0x1fc406b196bbf7,
	0x1fc3edb248cb62, 0x1fc3d2c43e593e, 0x1fc3b5de0591b5, 0x1fc396f599614d,
	0x1fc376005a4594, 0x1fc352f3069372, 0x1fc32dc1b2281b, 0x1fc3065fbd7888,
	0x1fc2dcbfcbf264, 0x1fc2b0d3b99fa0, 0x1fc2828c8ffcf0, 0x1fc251da79f164,
	0x1fc21eacb6d39e, 0x1fc1e8f18c6757, 0x1fc1b09637bb3d, 0x1fc17586dccd11,
	0x1fc137ae74d6b8, 0x1fc0f6f6bb2416, 0x1fc0b348184da4, 0x1fc06c898baff1,
	0x1fc022a092f365, 0x1fbfd5710f72b8, 0x1fbf84dd294890, 0x1fbf30c52fc60d,
	0x1fbed907770cc6, 0x1fbe7d80327ddc, 0x1fbe1e094ba615, 0x1fbdba7a354408,
	0x1fbd52a7b9f826, 0x1fbce663c6201b, 0x1fbc757d2c4de5, 0x1fbbffbf63b7aa,
	0x1fbb84f23fe6a2, 0x1fbb04d9a0d18e, 0x1fba7f351a70ad, 0x1fb9f3bf92b61a,
	0x1fb9622ed4abfc, 0x1fb8ca33174a18, 0x1fb82b76765b54, 0x1fb7859c5b895d,
	0x1fb6d840d55594, 0x1fb622f7d96943, 0x1fb5654c6f37e2, 0x1fb49ebfbf69d2,
	0x1fb3cec803e747, 0x1fb2f4cf539c40, 0x1fb21032442854, 0x1fb1203e5a9605,
	0x1fb0243042e1c3, 0x1faf1b31c479a7, 0x1fae045767e106, 0x1facde9dbf2d73,
	0x1faba8e640060b, 0x1faa61f399ff29, 0x1fa908656f66a2, 0x1fa79ab3508d3d,
	0x1fa61726d1f215, 0x1fa47bd48bea00, 0x1fa2c693c5c095, 0x1fa0f4f47df316,
	0x1f9f04336bbe0b, 0x1f9cf12b79f9bd, 0x1f9ab84415abc6, 0x1f98555b782fb9,
	0x1f95c3abd03f7a, 0x1f92fda9cef1f3, 0x1f8ffcda9ae41d, 0x1f8cb99e7385f8,
	0x1f892aec479608, 0x1f8545f904db90, 0x1f80fdc336039b, 0x1f7c427839e926,
	0x1f7700a3582ace, 0x1f71200f1a241d, 0x1f6a8234b7352c, 0x1f630000a8e267,
	0x1f5a66904fe3c6, 0x1f50724ece1173, 0x1f44c7665c6fdb, 0x1f36e5a38a59a4,
	0x1f26143450340b, 0x1f113e047b0414, 0x1ef6aefa57cbe7, 0x1ed38ca188151e,
	0x1ea2a61e122db1, 0x1e5961c78b267d, 0x1dddf62bac0bb1, 0x1cdb4dd9e4e8c0,
}
var we = [256]float64{
	9.655740063209187e-16, 7.08901424395524e-18, 1.1639412496691094e-17,
	1.5243915123532052e-17, 1.833284885723734e-17, 2.1089651094644777e-17,
	2.3611280778431305e-17, 2.5955957723108866e-17, 2.816173554197745e-17,
	3.0255041303213756e-17, 3.2255082548363685e-17, 3.417632340185021e-17,
	3.6029969787344476e-17, 3.7824907768696435e-17, 3.956832198097548e-17,
	4.1266117781759415e-17, 4.29232180844252e-17, 4.4543777432823665e-17,
	4.6131339814831816e-17, 4.768895725264631e-17, 4.9219280437279585e-17,
	5.0724629045031433e-17, 5.220704702792669e-17, 5.36683466171819e-17,
	5.5110143728350916e-17, 5.653388673239663e-17, 5.794088004852763e-17,
	5.933230365208939e-17, 6.070922932847175e-17, 6.207263431163189e-17,
	6.342341280303072e-17, 6.476238575956136e-17, 6.609030925769399e-17,
	6.740788167872716e-17, 6.871574991183808e-17, 7.001451473403925e-17,
	7.130473549660638e-17, 7.258693422414643e-17, 7.386159921381787e-17,
	7.512918820723722e-17, 7.63901311955082e-17, 7.764483290797843e-17,
	7.889367502729786e-17, 8.013701816675451e-17, 8.137520364041757e-17,
	8.260855505210033e-17, 8.383737972539134e-17, 8.506196999385318e-17,
	8.628260436784108e-17, 8.749954859216179e-17, 8.871305660690249e-17,
	8.992337142215353e-17, 9.113072591597904e-17, 9.233534356381783e-17,
	9.353743910649124e-17, 9.473721916312945e-17, 9.593488279457992e-17,
	9.713062202221516e-17, 9.832462230649506e-17, 9.951706298915067e-17,
	1.0070811770242944e-16, 1.0189795474846936e-16, 1.0308673745154215e-16,
	1.0427462448561881e-16, 1.0546177017945759e-16, 1.0664832480119145e-16,
	1.0783443482419484e-16, 1.0902024317583504e-16, 1.102058894705578e-16,
	1.1139151022861973e-16, 1.125772390816567e-16, 1.1376320696616842e-16,
	1.1494954230590088e-16, 1.1613637118402176e-16, 1.173238175059045e-16,
	1.185120031532669e-16, 1.1970104813034647e-16, 1.2089107070273853e-16,
	1.220821875294706e-16, 1.2327451378884152e-16, 1.2446816329851125e-16,
	1.2566324863028985e-16, 1.2685988122003978e-16, 1.2805817147307496e-16,
	1.2925822886541196e-16, 1.304601620412029e-16, 1.3166407890665726e-16,
	1.3287008672073811e-16, 1.3407829218289994e-16, 1.3528880151811755e-16,
	1.3650172055943978e-16, 1.377171548282881e-16, 1.3893520961270637e-16,
	1.4015599004375713e-16, 1.413796011702485e-16, 1.4260614803196652e-16,
	1.4383573573157902e-16, 1.4506846950536877e-16, 1.4630445479294757e-16,
	1.4754379730609514e-16, 1.4878660309686256e-16, 1.5003297862507367e-16,
	1.5128303082535392e-16, 1.5253686717381255e-16, 1.5379459575449967e-16,
	1.5505632532575771e-16, 1.5632216538658375e-16, 1.5759222624311761e-16,
	1.5886661907536842e-16, 1.6014545600429167e-16, 1.6142885015932787e-16,
	1.6271691574651305e-16, 1.6400976811727182e-16, 1.6530752383800372e-16,
	1.6661030076057423e-16, 1.679182180938229e-16, 1.6923139647620228e-16,
	1.7054995804966303e-16, 1.7187402653490321e-16, 1.7320372730810089e-16,
	1.7453918747925345e-16, 1.7588053597224919e-16, 1.772279036068007e-16,
	1.785814231823733e-16, 1.7994122956424645e-16, 1.8130745977185023e-16,
	1.826802530695253e-16, 1.8405975105985886e-16, 1.8544609777975702e-16,
	1.8683943979941934e-16, 1.8823992632438928e-16, 1.8964770930086177e-16,
	1.9106294352443775e-16, 1.9248578675252448e-16, 1.9391639982059002e-16,
	1.9535494676249099e-16, 1.9680159493510384e-16, 1.98256515147502e-16,
	1.997198817949343e-16, 2.0119187299787354e-16, 2.0267267074641993e-16,
	2.0416246105035898e-16, 2.0566143409519186e-16, 2.0716978440447378e-16,
	2.0868771100881602e-16, 2.1021541762192933e-16, 2.1175311282410767e-16,
	2.1330101025357798e-16, 2.148593288061664e-16, 2.1642829284376057e-16,
	2.1800813241207848e-16, 2.1959908346828715e-16, 2.2120138811904967e-16,
	2.2281529486961815e-16, 2.244410588846309e-16, 2.260789422613174e-16,
	2.2772921431586215e-16, 2.293921518837312e-16, 2.3106803963482143e-16,
	2.3275717040435356e-16, 2.344598455404959e-16, 2.361763752697775e-16,
	2.3790707908142777e-16, 2.3965228613186245e-16, 2.414123356706294e-16,
	2.4318757748922564e-16, 2.449783723943071e-16, 2.4678509270692897e-16,
	2.4860812278958527e-16, 2.5044785960295575e-16, 2.5230471329442175e-16,
	2.5417910782058127e-16, 2.5607148160617713e-16, 2.5798228824205314e-16,
	2.599119972249747e-16, 2.6186109474239247e-16, 2.638300845054943e-16,
	2.658194886341845e-16, 2.6782984859795257e-16, 2.6986172621694894e-16,
	2.719157047279819e-16, 2.7399238992058153e-16, 2.760924113487617e-16,
	2.782164236246436e-16, 2.8036510780069835e-16, 2.825391728480253e-16,
	2.847393572388174e-16, 2.8696643064198177e-16, 2.8922119574179956e-16,
	2.9150449019052937e-16, 2.9381718870700286e-16, 2.9616020533454657e-16,
	2.9853449587300453e-16, 3.009410605012618e-16, 3.033809466085003e-16,
	3.0585525185448604e-16, 3.08365127481531e-16, 3.1091178190342663e-16,
	3.1349648459966636e-16, 3.161205703467106e-16, 3.1878544382197136e-16,
	3.214925846206798e-16, 3.242435527309452e-16, 3.270399945182241e-16,
	3.2988364927722836e-16, 3.327763564171672e-16, 3.3572006335532446e-16,
	3.387168342045505e-16, 3.417688593525637e-16, 3.4487846604534244e-16,
	3.4804813010374423e-16, 3.51280488922298e-16, 3.5457835592247924e-16,
	3.579447366604277e-16, 3.613828468219061e-16, 3.648961323764543e-16,
	3.6848829220956213e-16, 3.721633036080208e-16, 3.7592545104162565e-16,
	3.797793587668875e-16, 3.837300278789214e-16, 3.877828785607896e-16,
	3.9194379843114294e-16, 3.9621919807867755e-16, 4.006160751056542e-16,
	4.0514208829565737e-16, 4.098056438903063e-16, 4.146159964290905e-16,
	4.1958336720733994e-16, 4.2471908418243855e-16, 4.3003574816674707e-16,
	4.355474314693952e-16, 4.4126991690360704e-16, 4.472209874259932e-16,
	4.534207798565834e-16, 4.598922204905932e-16, 4.666615664711476e-16,
	4.737590853262492e-16, 4.812199172829238e-16, 4.89085182739221e-16,
	4.97403423619194e-16, 5.06232507214416e-16, 5.156421828878083e-16,
	5.257175802022275e-16, 5.365640977112021e-16, 5.483144034258703e-16,
	5.611387454675159e-16, 5.752606481503331e-16, 5.909817641652101e-16,
	6.087231416180907e-16, 6.290979034877556e-16, 6.53049205356404e-16,
	6.821393079028929e-16, 7.192444966089362e-16, 7.706095350032097e-16,
	8.545517038584027e-16,
}
var fe = [256]float64{
	1, 0.9381436808621761, 0.9004699299257475, 0.8717043323812045,
	0.8477855006239904, 0.826993296643051, 0.8084216515230089,
	0.7915276369724962, 0.7759568520401161, 0.7614633888498967,
	0.7478686219851955, 0.7350380924314239, 0.7228676595935724,
	0.7112747608050763, 0.7001926550827885, 0.6895664961170783,
	0.6793505722647657, 0.6695063167319251, 0.660000841079,
	0.6508058334145713, 0.6418967164272663, 0.6332519942143663,
	0.6248527387036661, 0.6166821809152078, 0.6087253820796222,
	0.6009689663652324, 0.5934009016917337, 0.5860103184772683,
	0.5787873586028452, 0.571723048664826, 0.5648091929124005,
	0.5580382822625878, 0.5514034165406416, 0.54489823767244,
	0.5385168720028621, 0.5322538802630435, 0.52610421398362,
	0.5200631773682338, 0.5141263938147488, 0.5082897764106431,
	0.502549501841348, 0.49690198724154977, 0.49134386959403276,
	0.48587198734188514, 0.48048336393045443, 0.4751751930373776,
	0.4699448252839602, 0.4647897562504264, 0.45970761564213786,
	0.45469615747461567, 0.44975325116275516, 0.4448768734145487,
	0.44006510084235406, 0.4353161032156368, 0.43062813728845906,
	0.42599954114303457, 0.4214287289976168, 0.4169141864330031,
	0.41245446599716135, 0.40804818315203256, 0.40369401253053044,
	0.39939068447523124, 0.3951369818332903, 0.3909317369847973,
	0.3867738290841378, 0.38266218149600995, 0.3785957594095809,
	0.3745735676159022, 0.37059464843514606, 0.3666580797815142,
	0.3627629733548179, 0.3589084729487499, 0.35509375286678757,
	0.3513180164374835, 0.3475804946216372, 0.34388044470450263,
	0.3402171490667802, 0.33658991402867766, 0.33299806876180904,
	0.3294409642641363, 0.3259179723935562, 0.3224284849560891,
	0.3189719128449572, 0.3155476852271289, 0.31215524877417955,
	0.30879406693456013, 0.3054636192445902, 0.30216340067569347,
	0.29889292101558174, 0.2956517042812612, 0.2924392881618926,
	0.28925522348967775, 0.2860990737370769, 0.2829704145387808,
	0.2798688332369729, 0.27679392844851736, 0.27374530965280297,
	0.27072259679906, 0.2677254199320448, 0.26475341883506226,
	0.26180624268936303, 0.2588835497490163, 0.25598500703041543,
	0.2531102900156295, 0.25025908236886235, 0.24743107566532765,
	0.24462596913189213, 0.24184346939887724, 0.23908329026244915,
	0.23634515245705962, 0.23362878343743335, 0.2309339171696274,
	0.22826029393071667, 0.225607660116684, 0.22297576805812014,
	0.22036437584335944, 0.21777324714870044, 0.2152021510753786,
	0.2126508619929782, 0.21011915938898817, 0.20760682772422195,
	0.20511365629383763, 0.2026394390937089, 0.20018397469191118,
	0.19774706610509873, 0.1953285206795631, 0.1929281499767712,
	0.19054576966319528, 0.18818119940425418, 0.185834262762197,
	0.18350478709776735, 0.18119260347549615, 0.17889754657247814,
	0.17661945459049475, 0.17435816917135336, 0.17211353531531992,
	0.16988540130252752, 0.16767361861725005, 0.16547804187493587,
	0.16329852875190168, 0.1611349399175919, 0.15898713896931407,
	0.15685499236936512, 0.15473836938446797, 0.1526371420274428,
	0.15055118500103984, 0.1484803756438667, 0.14642459387834483,
	0.14438372216063466, 0.1423576454324721, 0.14034625107486234,
	0.13834942886358012, 0.13636707092642877, 0.13439907170221355,
	0.13244532790138744, 0.13050573846833072, 0.12858020454522814,
	0.12666862943751062, 0.12477091858083088, 0.12288697950954505,
	0.12101672182667474, 0.1191600571753276, 0.11731689921155547,
	0.11548716357863345, 0.11367076788274424, 0.11186763167005624,
	0.11007767640518532, 0.1083008254510337, 0.10653700405000158,
	0.1047861393065701, 0.10304816017125766, 0.10132299742595359,
	0.09961058367063709, 0.09791085331149216, 0.09622374255043278,
	0.09454918937605583, 0.09288713355604353, 0.09123751663104016,
	0.08960028191003284, 0.08797537446727019, 0.08636274114075689,
	0.0847623305323681, 0.08317409300963235, 0.08159798070923742,
	0.0800339475423199, 0.07848194920160644, 0.07694194317048052,
	0.07541388873405841, 0.07389774699236475, 0.07239348087570872,
	0.07090105516237181, 0.06942043649872875, 0.06795159342193662,
	0.06649449638533979, 0.06504911778675376, 0.06361543199980735,
	0.06219341540854101, 0.06078304644547963, 0.05938430563342025,
	0.05799717563120063, 0.05662164128374284, 0.055257689676697,
	0.05390531019604605, 0.052564494593071664, 0.051235237055126254,
	0.04991753428270636, 0.048611385573379476, 0.04731679291318154,
	0.046033761076175156, 0.04476229773294327, 0.043502413568888176,
	0.04225412241331622, 0.041017441380414806, 0.039792391023374105,
	0.03857899550307484, 0.03737728277295935, 0.03618728478193141,
	0.035009037697397404, 0.03384258215087432, 0.03268796350895953,
	0.031545232172893595, 0.030414443910466597, 0.029295660224637383,
	0.028188948763978622, 0.02709438378095579, 0.026012046645134207,
	0.024942026419731776, 0.023884420511558164, 0.02283933540638523,
	0.02180688750428357, 0.020787204072578114, 0.01978042433800974,
	0.018786700744696024, 0.017806200410911355, 0.01683910682603994,
	0.015885621839973156, 0.014945968011691148, 0.014020391403181943,
	0.013109164931254991, 0.012212592426255378, 0.0113310135978346,
	0.01046481018102998, 0.009614413642502213, 0.008780314985808977,
	0.007963077438017043, 0.007163353183634991, 0.006381905937319183,
	0.005619642207205488, 0.0048776559835424, 0.0041572951208338005,
	0.003460264777836907, 0.0027887987935740783, 0.002145967743718907,
	0.0015362997803015726, 0.0009672692823271743, 0.0004541343538414966,
}
